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


def test_autodetect_fallback_single_project(test_env, tmp_path_factory, monkeypatch):
    # 1. Setup
    # Initialize the "test-app" project using the test_env fixture.
    # test_env.setup() already ensures "test-app" is the only project in its managed dir.
    init_result = subprocess.run(
        ["devbox", "init", str(test_env.manifest_repo), "--name", "test-app", "--branch", "master"],
        capture_output=True,
        text=True,
        check=True  # Fail if init fails
    )
    assert "Project has been successfully initialized!" in init_result.stdout
    original_project_name = "test-app"

    # Create a temporary directory outside the project and cd into it.
    non_project_cwd = tmp_path_factory.mktemp("non_project_cwd")
    monkeypatch.chdir(non_project_cwd)

    # 2. Execution
    # Run `devbox info` from the non-project directory.
    # It should autodetect the "test-app" project.
    info_result = subprocess.run(
        ["devbox", "info"],
        capture_output=True,
        text=True,
        check=True  # Fail if devbox info returns non-zero
    )

    # 3. Assertion
    # Assert that the command executed successfully (implicit from check=True).
    # Assert that the output indicates it's operating on the correct project.
    # Assuming `devbox info` output is plain text and includes the project name.
    # If it's JSON, parsing would be better. For now, simple string check.
    assert f"Name: {original_project_name}" in info_result.stdout
    # A more robust check might involve looking for a specific path or ID if available.
    # For example, if it prints the project working directory:
    expected_project_dir_info = f"Working dir: {str(test_env.project_dir)}"
    assert expected_project_dir_info in info_result.stdout

    # 4. Cleanup
    # test_env fixture handles cleanup of the project.
    # tmp_path_factory and monkeypatch handle their respective cleanups.