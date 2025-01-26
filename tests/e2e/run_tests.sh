#!/bin/bash

# Usage:
# ./run_tests.sh -v  # Run with verbose output
# ./run_tests.sh -k test_info  # Run only tests with "test_info" in their name
# ./run_tests.sh -s  # Show print statements during test execution

# Colors for better output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Setting up Python virtual environment...${NC}"

# Create and activate venv if it doesn't exist
if [ ! -d "venv" ]; then
    python3 -m venv venv
fi

source venv/bin/activate

echo -e "${GREEN}Installing dependencies...${NC}"
pip install -r requirements.txt

echo -e "${GREEN}Running tests...${NC}"
pytest "$@"

# Deactivate venv
deactivate 