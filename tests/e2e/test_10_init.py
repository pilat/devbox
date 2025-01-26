import subprocess
from pathlib import Path

def test_project_initialization(test_env):
    # Test initialization
    result = subprocess.run(
        ["devbox", "init", str(test_env.manifest_repo), "--name", "test-app", "--branch", "master"],
        capture_output=True,
        text=True
    )
    assert "Project has been successfully initialized!" in result.stdout

    # Test project structure
    for file in ["docker-compose.yml", ".devboxstate", ".env"]:
        assert (test_env.project_dir / file).exists(), f"File {file} should exist"

    # Test git exclude configuration
    exclude_file = test_env.project_dir / ".git/info/exclude"
    expected_content = "/sources/\n/.devboxstate\n/.env"
    assert exclude_file.read_text().strip() == expected_content

    # Test project list
    result = subprocess.run(
        ["devbox", "list"],
        capture_output=True,
        text=True
    )
    assert "test-app" in result.stdout 