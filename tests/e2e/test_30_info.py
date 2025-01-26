import subprocess

def test_info_operations(test_env):
    # Initialize project
    subprocess.run(
        ["devbox", "init", str(test_env.manifest_repo), "--name", "test-app", "--branch", "master"],
        capture_output=True
    )

    # Test info after init should fail
    result = subprocess.run(
        ["devbox", "info", "--name", "test-app"],
        capture_output=True,
        text=True
    )
    assert "No such file or directory" in result.stdout + result.stderr

    # Test updating sources
    result = subprocess.run(
        ["devbox", "update", "--name", "test-app"],
        capture_output=True,
        text=True
    )
    assert "Source service-1  Synced" in result.stdout

    # Test info after update
    result = subprocess.run(
        ["devbox", "info", "--name", "test-app"],
        capture_output=True,
        text=True
    )
    assert "Initial commit" in result.stdout

    # Test services are not started
    result = subprocess.run(
        ["docker", "ps", "--filter", "name=test-app"],
        capture_output=True,
        text=True
    )
    assert "Up" not in result.stdout 