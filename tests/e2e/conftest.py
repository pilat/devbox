import os
import shutil
import tempfile
import subprocess
from pathlib import Path
import pytest

class TestEnv:
    def __init__(self):
        self.home = Path.home()
        self.project_dir = self.home / ".devbox/test-app"
        self.temp_dir = None
        self.manifest_repo = None
        self.manifest_dir = None
        self.source_1_repo = None
        self.source_1_dir = None
        self.fixtures_dir = Path(__file__).parent / "fixtures"

    def setup(self):
        self.temp_dir = Path(tempfile.mkdtemp(prefix="devbox-e2e-"))
        self.manifest_repo = self.temp_dir / "manifest.git"
        self.manifest_dir = self.temp_dir / "manifest"
        self.source_1_repo = self.temp_dir / "source-1.git"
        self.source_1_dir = self.temp_dir / "source-1"

        # Clean previous project if exists
        if self.project_dir.exists():
            shutil.rmtree(self.project_dir)

        # Setup manifest repo
        subprocess.run(["git", "init", "--bare", str(self.manifest_repo)], capture_output=True)
        self.manifest_dir.mkdir()
        os.chdir(str(self.manifest_dir))
        subprocess.run(["git", "init"], capture_output=True)

        # Copy and configure manifest files
        for file in self.fixtures_dir.glob("manifest/*"):
            shutil.copy(file, self.manifest_dir)
        
        # Update docker-compose.yml with correct repo path
        compose_file = self.manifest_dir / "docker-compose.yml"
        content = compose_file.read_text()
        content = content.replace("SOURCE_1_REPO", str(self.source_1_repo))
        compose_file.write_text(content)

        # Push manifest
        subprocess.run(["git", "remote", "add", "origin", str(self.manifest_repo)], capture_output=True)
        subprocess.run(["git", "add", "."], capture_output=True)
        subprocess.run(["git", "commit", "-m", "Initial commit"], capture_output=True)
        subprocess.run(["git", "push", "origin", "master"], capture_output=True)

        # Setup source 1 repo
        subprocess.run(["git", "init", "--bare", str(self.source_1_repo)], capture_output=True)
        self.source_1_dir.mkdir()
        os.chdir(str(self.source_1_dir))
        subprocess.run(["git", "init"], capture_output=True)

        # Copy service-1 files
        for item in (self.fixtures_dir / "service-1").iterdir():
            if item.is_dir():
                shutil.copytree(item, self.source_1_dir / item.name)
            else:
                shutil.copy(item, self.source_1_dir)

        # Push source 1
        subprocess.run(["git", "remote", "add", "origin", str(self.source_1_repo)], capture_output=True)
        subprocess.run(["git", "add", "."], capture_output=True)
        subprocess.run(["git", "commit", "-m", "Initial commit"], capture_output=True)
        subprocess.run(["git", "push", "origin", "master"], capture_output=True)

    def cleanup(self):
        # Stop and remove containers
        subprocess.run(
            ["docker", "ps", "-aq", "--filter", "name=test-app"],
            capture_output=True,
            text=True
        ).stdout.strip().split('\n')

        containers = [c for c in subprocess.run(
            ["docker", "ps", "-aq", "--filter", "name=test-app"],
            capture_output=True,
            text=True
        ).stdout.strip().split('\n') if c]

        for container in containers:
            subprocess.run(["docker", "stop", "-t0", container], capture_output=True)
            subprocess.run(["docker", "rm", "-f", container], capture_output=True)

        # Remove image
        subprocess.run(["docker", "image", "rm", "local/service-1"], capture_output=True)

        # Clean directories
        if self.project_dir.exists():
            shutil.rmtree(self.project_dir)
        if self.temp_dir and self.temp_dir.exists():
            shutil.rmtree(self.temp_dir)

@pytest.fixture
def test_env():
    env = TestEnv()
    env.setup()
    yield env
    env.cleanup()
