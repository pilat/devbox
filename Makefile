NAME := devbox
VERSION ?= $(shell git describe --tags || echo "unknown")
GOBUILD = CGO_ENABLED=0 go build -trimpath -ldflags "-s -w"

# Release
PLATFORM_LIST = \
        darwin-amd64 \
        darwin-arm64 \
        linux-amd64 \
        linux-arm64

darwin-%:
	GOARCH=$* GOOS=darwin $(GOBUILD) -o $(NAME)_$(VERSION)_$@/$(NAME) ./cmd/devbox/...

linux-%:
	GOARCH=$* GOOS=linux $(GOBUILD) -o $(NAME)_$(VERSION)_$@/$(NAME) ./cmd/devbox/...

gz_releases=$(addsuffix .tar.gz, $(PLATFORM_LIST))
$(gz_releases): %.tar.gz : %
	tar czf $(NAME)_$(VERSION)_$@ -C $(NAME)_$(VERSION)_$</ ../LICENSE $(NAME)
	sha256sum $(NAME)_$(VERSION)_$@ > $(NAME)_$(VERSION)_$@.sha256

.PHONY: releases
releases: $(gz_releases)
