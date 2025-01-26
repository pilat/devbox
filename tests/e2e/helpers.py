import time
import subprocess
import requests

def wait_for(condition_func, timeout=30, interval=1):
    """
    Wait for a condition to be true within a timeout period.
    
    Args:
        condition_func: Function that returns True when condition is met
        timeout: Maximum time to wait in seconds
        interval: Time between checks in seconds
    
    Returns:
        bool: True if condition was met, False if timeout occurred
    """
    start_time = time.time()
    while True:
        try:
            if condition_func():
                return True
        except:
            pass
        
        if time.time() - start_time >= timeout:
            return False
        time.sleep(interval)

def check_containers_up(count=2):
    """Check if specified number of test-app containers are running."""
    result = subprocess.run(
        ["docker", "ps", "--filter", "name=test-app"],
        capture_output=True,
        text=True
    )
    return result.stdout.count("Up") == count

def check_containers_down():
    """Check if no test-app containers are running."""
    result = subprocess.run(
        ["docker", "ps", "--filter", "name=test-app"],
        capture_output=True,
        text=True
    )
    return "Up" not in result.stdout

def check_service_response(url, expected_text):
    """Check if a service responds with expected text."""
    try:
        response = requests.get(url)
        return expected_text in response.text
    except:
        return False
    
def remove_build_context_from_docker_compose(compose_file):
    content = compose_file.read_text()
    content = content.replace("""    build:
      context: ./sources/service-1
      dockerfile: Dockerfile""", "")
    content = content.replace("""image: local/service-1:latest""", """image: golang:1.21-alpine""")
    compose_file.write_text(content)

def remove_volume_from_docker_compose(compose_file):
    content = compose_file.read_text()
    content = content.replace("""    volumes:
      - ./sources/service-1:/app""", "")
    compose_file.write_text(content)
