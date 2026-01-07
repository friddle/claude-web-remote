# ClauDED Build Guide

## Quick Start

### Build for Current Platform
```bash
make build
```

### Build for All Platforms (Recommended)
```bash
make build-all
```

This will build all platforms and automatically create tar.gz packages in the `output/` directory.

## Supported Platforms

| Platform | Architecture | Binary | Package |
|----------|-------------|--------|---------|
| Linux | AMD64 | `clauded-linux-amd64` | `clauded-{version}-linux-amd64.tar.gz` |
| Linux | ARM64 | `clauded-linux-arm64` | `clauded-{version}-linux-arm64.tar.gz` |
| macOS | AMD64 (Intel) | `clauded-darwin-amd64` | `clauded-{version}-darwin-amd64.tar.gz` |
| macOS | ARM64 (Apple Silicon) | `clauded-darwin-arm64` | `clauded-{version}-darwin-arm64.tar.gz` |

## Build Outputs

All build outputs are in the `output/` directory:

```
output/
├── clauded-linux-amd64
├── clauded-linux-arm64
├── clauded-darwin-amd64
├── clauded-darwin-arm64
├── clauded-{version}-linux-amd64.tar.gz
├── clauded-{version}-linux-arm64.tar.gz
├── clauded-{version}-darwin-amd64.tar.gz
└── clauded-{version}-darwin-arm64.tar.gz
```

**Note**: The `output/` directory is in `.gitignore`.

## Makefile Targets

### Building
- `make build` - Build for current platform
- `make build-all` - Build all platforms and create packages (recommended)
- `make build-linux` - Build Linux versions only
- `make build-darwin` - Build macOS versions only

### Packaging
- `make package` - Package already-built binaries (requires binaries to exist)

### Utilities
- `make clean` - Remove output directory
- `make info` - Show build configuration
- `make help` - Display help information

## Examples

```bash
# Build all platforms (automatically packages)
make build-all

# Clean and rebuild
make clean && make build-all

# Check build info
make info

# View help
make help
```

## Installation

### From Binary
```bash
# Extract and install
tar -xzf output/clauded-{version}-linux-amd64.tar.gz
sudo cp clauded-linux-amd64 /usr/local/bin/clauded
sudo chmod +x /usr/local/bin/clauded
```

### From Package Directly
```bash
# Extract and run
tar -xzf output/clauded-{version}-linux-amd64.tar.gz
./clauded-linux-amd64 --help
```

## Version Information

Builds include version information:
- Version: Git tag or commit hash
- Build time: UTC timestamp
- Git commit: Short commit hash

View version info:
```bash
make info
```

## Tips

1. **Use `make build-all`** - This builds all platforms and creates packages automatically
2. **Output in `output/`** - All binaries and packages are in this directory
3. **Clean build** - Use `make clean && make build-all` for a fresh build
4. **Check packages** - Verify package contents with `tar -tzf output/clauded-*.tar.gz`
