package planner

const (
	provisionerImageName  = "local/gh_provisioner:latest"
	provisionerDockerfile = `FROM docker.io/bitnami/git:2.47.1

COPY <<'EOF' /usr/local/bin/entrypoint.sh
#!/bin/bash

set -e

if [ -n "${GITHUB_TOKEN}" ]; then
	echo "machine github.com" > ~/.netrc
	echo "login whatever" >> ~/.netrc
	echo "password ${GITHUB_TOKEN}" >> ~/.netrc
	chmod 600 ~/.netrc
fi

REPO_DIR="/workspace/${TARGET_FOLDER}"

# Configure git safe directory
git config --global --add safe.directory "*"

# Clone or update the repository
if [ -d "$REPO_DIR" ]; then
    rm -f "${REPO_DIR}/.git/index.lock" || true
    cd "$REPO_DIR"
    git reset --hard
    git clean -fd
else
    if ! git clone --no-checkout --depth 1 "$REPO_URL" "$REPO_DIR"; then
        exit 100
    fi
    cd "$REPO_DIR"
fi

# Handle sparse checkout
if [ -n "$SPARSE_CHECKOUT" ]; then
    git sparse-checkout init --cone

	IFS=',' read -r -a ITEMS <<< "$SPARSE_CHECKOUT"
	COMMAND="git sparse-checkout set"
	for ITEM in "${ITEMS[@]}"; do
		COMMAND+=" \"$ITEM\""
	done
	eval "$COMMAND"
    # git sparse-checkout set "$SPARSE_CHECKOUT"
else
    git sparse-checkout disable
fi

git checkout "${BRANCH_NAME}"

git log -1 --pretty=format:"%H%n%an%n%ad%n%s"
EOF

RUN chmod +x /usr/local/bin/entrypoint.sh
`
)
