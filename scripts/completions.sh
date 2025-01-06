#!/bin/sh
set -e

rm -rf completions || true
mkdir completions

for sh in bash zsh fish; do
	go run ./cmd/devbox/... completion "$sh" >"completions/devbox.$sh"
done
