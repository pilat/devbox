#!/bin/sh
set -e

go mod vendor

rm -f ./vendor/github.com/docker/compose/v2/pkg/watch/watcher_darwin.go
sed -i.bak '/build.*darwin/d' ./vendor/github.com/docker/compose/v2/pkg/watch/watcher_naive.go
rm -f ./vendor/github.com/docker/compose/v2/pkg/watch/watcher_naive.go.bak
