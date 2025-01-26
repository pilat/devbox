import subprocess
import os
from pathlib import Path
import time
import helpers


def test_project_context_detection(test_env):
    # Initialize and update project
    subprocess.run(
        ["devbox", "init", str(test_env.manifest_repo), "--name", "test-app", "--branch", "master"],
        capture_output=True
    )
    subprocess.run(
        ["devbox", "update", "--name", "test-app"],
        capture_output=True
    )

    # Test info in random directory not working
    os.chdir("/tmp")
    result = subprocess.run(
        ["devbox", "info"],
        capture_output=True,
        text=True
    )
    assert "Error has occurred" in result.stdout + result.stderr

    # Test info in random directory works if we mention project name
    os.chdir("/tmp")
    result = subprocess.run(
        ["devbox", "info", "--name", "test-app"],
        capture_output=True,
        text=True
    )
    assert "Initial commit" in result.stdout

    # Test info in source directory
    os.chdir(str(test_env.source_1_dir))
    result = subprocess.run(
        ["devbox", "info"],
        capture_output=True,
        text=True
    )
    assert "Initial commit" in result.stdout

    # Test info in subdirectory
    os.chdir(str(test_env.source_1_dir / "cmd" / "service-1"))
    result = subprocess.run(
        ["devbox", "info"],
        capture_output=True,
        text=True
    )
    assert "Initial commit" in result.stdout
    
    # Test that we can detect project just because it mentioned in project's sources
    helpers.remove_volume_from_docker_compose(test_env.project_dir / "docker-compose.yml")
    os.chdir(str(test_env.source_1_dir / "cmd" / "service-1"))
    result = subprocess.run(
        ["devbox", "info"],
        capture_output=True,
        text=True
    )
    assert "Initial commit" in result.stdout