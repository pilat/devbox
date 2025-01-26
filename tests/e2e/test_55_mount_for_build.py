import subprocess
import os
from helpers import wait_for, check_containers_up, check_service_response, remove_volume_from_docker_compose

def test_mount_for_build_operations(test_env):
    # Initialize and update project
    subprocess.run(
        ["devbox", "init", str(test_env.manifest_repo), "--name", "test-app", "--branch", "master"],
        capture_output=True
    )
    subprocess.run(
        ["devbox", "update", "--name", "test-app"],
        capture_output=True
    )

    remove_volume_from_docker_compose(test_env.project_dir / "docker-compose.yml")

    # Check services are not started
    result = subprocess.run(
        ["docker", "ps", "--filter", "name=test-app"],
        capture_output=True,
        text=True
    )
    assert "Up" not in result.stdout, "Services should not be started"

    # Test mounting source
    os.chdir(str(test_env.source_1_dir))
    result = subprocess.run(
        ["devbox", "mount"], #, "--source", "./sources/service-1"],
        capture_output=True,
        text=True
    )
    assert "LOCAL PATH" in result.stdout
    assert "./sources/service-1" in result.stdout

    # Start project
    subprocess.run(
        ["devbox", "up", "--name", "test-app"],
        capture_output=True
    )

    # Wait for containers to be up
    assert wait_for(check_containers_up), "Project containers should be running"

    # Test info shows commit
    result = subprocess.run(
        ["devbox", "info"],
        capture_output=True,
        text=True
    )
    assert "Initial commit" in result.stdout

    # Test code changes
    main_go = test_env.source_1_dir / "cmd" / "service-1" / "main.go"
    content = main_go.read_text()
    new_content = content.replace(
        "Hello, World from service 1",
        "Hello, World from service 1 updated"
    )
    main_go.write_text(new_content)

    # Restart service
    subprocess.run(
        ["devbox", "restart", "service-1"],
        capture_output=True
    )

    # Verify changes
    assert wait_for(
        lambda: check_service_response("http://localhost:8081", "Hello, World from service 1 updated")
    ), "Service should return updated response"
