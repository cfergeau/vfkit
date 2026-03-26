# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

vfkit is a command-line tool for starting and managing virtual machines on macOS using Apple's Virtualization Framework. It provides both a CLI tool (`vfkit`) and a Go package (`github.com/crc-org/vfkit/pkg/config`) that can be used by other applications to programmatically create and manage VMs.

**Critical Platform Requirements:**
- macOS only - all code uses `//go:build darwin` build tags
- Requires macOS 12+ (some features require macOS 13+)
- Supports both Intel (amd64) and Apple Silicon (arm64) architectures
- Uses CGO extensively via the Code-Hex/vz bindings to Apple's Virtualization Framework

## Build and Development Commands

**Building:**
```bash
make                    # Builds universal binary (amd64+arm64) at out/vfkit
make build             # Same as above
```

The build process:
- Builds separate binaries for amd64 and arm64
- Codesigns both binaries (required for Virtualization Framework)
- Uses `lipo` to create a universal binary
- Version automatically embedded by Go 1.24+ from git VCS information

**Testing:**
```bash
make test              # Run all tests (unit + integration)
make test-unit         # Unit tests: go test -v ./pkg/... ./cmd/vfkit/...
make test-integration  # Integration tests: go test -v -timeout 20m ./test
```

Integration tests require macOS and may create actual VMs. They have a 20-minute timeout.

**Linting:**
```bash
make lint              # Run golangci-lint (installs to tools/bin/ if needed)
```

Linters enabled: gocritic, gosec, misspell, revive, errcheck, unused. Configuration in `.golangci.yml`.

**Running a single test:**
```bash
go test -v -run TestName ./pkg/package/...
```

**Cleaning:**
```bash
make clean             # Remove out/ directory
```

## Code Architecture

### Package Structure

- **`cmd/vfkit/`** - Main CLI application entry point
  - Uses cobra for command structure
  - Parses command-line args and creates VM config
  - Handles signals, exit handlers, and VM lifecycle

- **`pkg/config/`** - VM configuration types and command-line generation
  - **Critical**: This package is designed to be cross-compilable (doesn't import Code-Hex/vz directly)
  - Provides `VirtualMachine` type describing VM configuration (CPUs, memory, bootloader, devices)
  - Can generate vfkit command-line from config via `ToCmdLine()`
  - Can parse command-line args into config via `FromOptions()`
  - Used by applications that want to start VMs without directly using vfkit CLI
  - Subtypes: `Bootloader` implementations, `VirtioDevice` implementations

- **`pkg/vf/`** - Virtualization Framework wrappers
  - Creates actual VZ objects from config types
  - Implements device creation (virtio-blk, virtio-net, virtio-serial, etc.)
  - Handles bootloader setup (Linux, EFI, macOS)
  - Contains architecture-specific code (_arm64.go, _amd64.go files)
  - Rosetta support (arm64 only)

- **`pkg/cmdline/`** - Command-line parsing
  - Defines `Options` struct for parsed CLI arguments
  - Handles version information

- **`pkg/rest/`** - REST API server
  - Optional HTTP/Unix socket API for controlling running VMs
  - Endpoints: `/vm/state`, `/vm/inspect`
  - Can start/stop/pause/resume VMs
  - Uses gin framework

- **`pkg/util/`** - Shared utilities
  - Exit handlers
  - String helpers

- **`pkg/process/`** - Process management utilities

- **`test/`** - Integration tests
  - Tests actual VM creation and lifecycle
  - Platform-specific helpers (`osprovider.go`, `osversion.go`)
  - Requires real macOS environment

### Key Design Patterns

**Config → VF Separation:**
The `pkg/config` package defines "what to run" while `pkg/vf` handles "how to run it with Virtualization Framework". This allows `pkg/config` to be cross-compiled and used on non-macOS systems for generating vfkit command lines.

**VMComponent Interface:**
Both devices and bootloaders implement the `VMComponent` interface with `FromOptions()` and `ToCmdLine()` methods, enabling bidirectional conversion between command-line args and typed config objects.

**Architecture-Specific Code:**
Files with `_amd64.go` or `_arm64.go` suffixes contain platform-specific implementations. The Rosetta device only exists on arm64.

## Testing Considerations

- Integration tests create actual VMs and require significant resources
- Tests may need specific OS versions for certain features (e.g., EFI requires macOS 13+)
- Some tests use assets in `test/assets/` directory
- When modifying VM creation code, always run integration tests

## Common Development Patterns

**Adding a new device type:**
1. Add device struct in `pkg/config/virtio.go`
2. Implement `VMComponent` interface (FromOptions, ToCmdLine)
3. Add JSON marshaling/unmarshaling in `pkg/config/json.go`
4. Implement VZ device creation in `pkg/vf/virtio.go`
5. Add tests in `pkg/config/virtio_test.go` and `pkg/vf/`
6. Update documentation in `doc/usage.md`

**Adding a new bootloader:**
1. Add bootloader struct in `pkg/config/bootloader.go`
2. Implement `Bootloader` interface
3. Add JSON support in `pkg/config/json.go`
4. Implement VZ bootloader creation in `pkg/vf/bootloader.go`
5. Add tests

## Important Notes

- **CGO Required**: All builds require CGO_ENABLED=1 and proper SDK paths
- **Codesigning Required**: Binaries must be codesigned to use Virtualization Framework (done automatically by Makefile)
- **Minimum macOS Version**: Set via CGO_CFLAGS=-mmacosx-version-min=13.0
- **Kernel Requirements**: On Apple Silicon, Linux kernels must be uncompressed when using the Linux bootloader (not required with EFI bootloader)
- **REST API**: Disabled by default, enable with `--restful-uri`
- **Time Sync**: Requires qemu-guest-agent in the guest VM

## Dependencies

Key external dependencies:
- `github.com/Code-Hex/vz/v3` - Virtualization Framework Go bindings
- `github.com/spf13/cobra` - CLI framework
- `github.com/gin-gonic/gin` - REST API framework
- `go.podman.io/common` - Common utilities (including strongunits for memory sizes)

## Documentation

- `doc/usage.md` - Comprehensive CLI usage and all device/bootloader options
- `doc/quickstart.md` - Quick start guide with examples
- `doc/supported_platforms.md` - Platform support matrix
- `README.md` - Project overview and installation
