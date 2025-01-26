import subprocess
from helpers import wait_for, check_containers_up, check_containers_down, check_service_response

def test_project_operations(test_env):
    # Initialize project
    subprocess.run(
        ["devbox", "init", str(test_env.manifest_repo), "--name", "test-app", "--branch", "master"],
        capture_output=True
    )

    # Start project
    subprocess.run(
        ["devbox", "up", "--name", "test-app"],
        capture_output=True
    )

    # Wait for containers to be up
    assert wait_for(check_containers_up), "Project containers should be running"

    # Stop project
    subprocess.run(
        ["devbox", "down", "--name", "test-app"],
        capture_output=True
    )

    # Wait for containers to stop
    assert wait_for(check_containers_down), "Project containers should be stopped"
