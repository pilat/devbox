import subprocess

def test_project_cleanup(test_env):
    # Initialize and start project
    subprocess.run(
        ["devbox", "init", str(test_env.manifest_repo), "--name", "test-app", "--branch", "master"],
        capture_output=True
    )
    subprocess.run(
        ["devbox", "up", "--name", "test-app"],
        capture_output=True
    )

    # Test project destroy
    subprocess.run(
        ["devbox", "destroy", "--name", "test-app"],
        capture_output=True
    )
    assert not test_env.project_dir.exists(), "Project directory should be removed"

    # Verify no containers
    result = subprocess.run(
        ["docker", "ps", "-q", "--filter", "name=test-app"],
        capture_output=True,
        text=True
    )
    assert not result.stdout.strip(), "No containers should be running"

    # Make sure files were removed
    assert not test_env.project_dir.exists(), "Project files should be removed" 