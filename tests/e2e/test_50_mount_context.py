import subprocess
import os
from helpers import wait_for, check_containers_up, check_service_response
import time
import helpers


def test_mount_context_detection(test_env):
    subprocess.run(
        ["devbox", "init", str(test_env.manifest_repo), "--name", "test-app", "--branch", "master"],
        capture_output=True
    )
    subprocess.run(
        ["devbox", "update", "--name", "test-app"],
        capture_output=True
    )

    # Test mount in random directory not working even if we mention project name
    os.chdir("/tmp")
    result = subprocess.run(
        ["devbox", "mount", "--name", "test-app"],
        capture_output=True,
        text=True
    )
    assert "Error has occurred" in result.stdout + result.stderr

    # Test mount in project directory
    os.chdir(str(test_env.source_1_dir))
    result = subprocess.run(
        ["devbox", "mount"],
        capture_output=True,
        text=True
    )
    assert "./sources/service-1" in result.stdout

    # We have to umount
    subprocess.run(
        ["devbox", "umount"],
        capture_output=True,
        text=True
    )

    # It is expected that mount going to work even if we are in subdir
    os.chdir(str(test_env.source_1_dir / "cmd" / "service-1"))
    result = subprocess.run(
        ["devbox", "mount"],
        capture_output=True,
        text=True,
        env={
            **os.environ,
            "DEVBOX_DEBUG": "true",
        }
    )
    assert "./sources/service-1" in result.stdout

    # TODO: test mount from subdir when repo is a part of sparse checkout

    # Test that it will work if not mentioned in volumes and mentioned in build.context
    helpers.remove_volume_from_docker_compose(test_env.project_dir / "docker-compose.yml")
    os.chdir(str(test_env.source_1_dir))
    result = subprocess.run(
        ["devbox", "mount"],
        capture_output=True,
        text=True
    )
    assert "./sources/service-1" in result.stdout
