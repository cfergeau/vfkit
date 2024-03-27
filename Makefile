.PHONY: all build clean test

GIT_VERSION ?= $(shell git describe --always --dirty)
CGO_CFLAGS=-mmacosx-version-min=11.0
VERSION_LDFLAGS=-X github.com/crc-org/vfkit/pkg/cmdline.gitVersion=$(GIT_VERSION)

all: build

TOOLS_DIR := tools
include tools/tools.mk

build: bin/vfkit

test: build
	@go test -v ./pkg/...
	@go test -v -timeout 20m ./test

clean:
	rm -rf bin

bin/vfkit-amd64 bin/vfkit-arm64: bin/vfkit-%: force-build
	@mkdir -p $(@D)
	CGO_ENABLED=1 CGO_CFLAGS=$(CGO_CFLAGS) GOOS=darwin GOARCH=$* go build -ldflags "$(VERSION_LDFLAGS)" -o $@ ./cmd/vfkit
	codesign -f --entitlements vf.entitlements -s - $@

bin/vfkit: bin/vfkit-amd64 bin/vfkit-arm64
	cd $(@D) && lipo -create $(^F) -output $(@F)

# the go compiler is doing a good job at not rebuilding unchanged files
# this phony target ensures bin/vfkit-* are always considered out of date
# and rebuilt. If the code was unchanged, go won't rebuild anything so that's
# fast. Forcing the rebuild ensure we rebuild when needed, ie when the source code
# changed, without adding explicit dependencies to the go files/go.mod
.PHONY: force-build
force-build:

.PHONY: lint
lint: $(TOOLS_BINDIR)/golangci-lint
	"$(TOOLS_BINDIR)"/golangci-lint run
