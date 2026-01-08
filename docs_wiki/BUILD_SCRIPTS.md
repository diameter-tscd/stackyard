# Enhanced Build Scripts with Garble Obfuscation

## What's New in This Version

### Enhanced Build Scripts (Latest Update)
The build scripts have been significantly enhanced with the following new features:

- **Automatic Tool Installation**: Scripts now automatically check for and install required Go tools (`goversioninfo`, `garble`) if not present
- **Interactive Obfuscation Choice**: Users are prompted to choose between standard Go build or garble obfuscation build
- **Timeout Handling**: 10-second timeout with sensible default (no obfuscation) for CI/CD compatibility
- **Updated Step Numbers**: Progress indicators now show [0/6] through [5/6] to account for tool checking phase
- **Enhanced Error Handling**: Better error messages and recovery options for tool installation failures

### TUI Improvements
The Terminal User Interface has been significantly enhanced with advanced features:

- **Scrollable Log Display**: Full scrolling support through application logs with arrow keys, page up/down, and home/end
- **Real-time Log Filtering**: Press "/" to open a modal filter dialog for searching logs by content or level
- **Auto-scroll Control**: Manual toggle (F1) to enable/disable automatic scrolling to bottom on new logs
- **Log Management**: Press F2 to clear all accumulated logs and reset the view state
- **Modal Filter Interface**: Clean, centered dialog with black background for focused log filtering
- **Sticky Header/Footer**: App information and controls remain visible while scrolling through logs
- **Thread-safe Operations**: All log operations are properly synchronized for concurrent access
- **Unlimited Log Storage**: Removed 1000 log limit to allow unlimited log storage
- **Default Auto-scroll**: Auto-scroll enabled by default on application startup
- **Reusable Dialog System**: Template-based dialog components in `pkg/tui/template/` for easy reuse
- **Dialog Footer Cleanup**: Removed scroll count from footer for cleaner interface

### Customizable Parameter Parsing System
The application now features a highly customizable parameter parsing system that allows easy addition and configuration of command-line flags:

- **Flag Definition System**: Command-line flags are defined in `cmd/app/main.go` using a structured `FlagDefinition` type
- **Dynamic Parsing**: The parsing logic in `pkg/utils/parameter.go` automatically handles flag validation and type conversion
- **Easy Extension**: Add new flags by simply adding entries to the flagDefinitions slice
- **Type Safety**: Support for string, int, and bool flag types with built-in validation
- **Custom Validators**: Each flag can have custom validation functions for business logic rules
- **Modular Architecture**: Parsing logic separated from configuration for maintainability

**Example of Adding a New Flag:**
```go
// In cmd/app/main.go
var flagDefinitions = []utils.FlagDefinition{
    {
        Name:         "c",
        DefaultValue: "",
        Description:  "URL to load configuration from (YAML format)",
        Validator: func(value interface{}) error {
            if str, ok := value.(string); ok && str != "" {
                if _, err := url.ParseRequestURI(str); err != nil {
                    return fmt.Errorf("invalid config URL format: %w", err)
                }
            }
            return nil
        },
    },
    // Add new flags easily
    {
        Name:         "port",
        DefaultValue: 8080,
        Description:  "Server port to listen on",
        Validator: func(value interface{}) error {
            if port, ok := value.(int); ok && (port < 1 || port > 65535) {
                return fmt.Errorf("port must be between 1 and 65535")
            }
            return nil
        },
    },
}
```

### Previous Features (Still Included)
- Cross-platform support (Unix/Linux/macOS and Windows)
- Automatic backup and archiving of previous builds
- Process management (stops running application instances)
- Asset copying (config files, web assets, databases)
- Comprehensive error handling and troubleshooting

## Overview

The enhanced build scripts (`scripts/build.sh` for Unix/Linux/macOS and `scripts/build.bat` for Windows) now include automatic tool installation and user choice for code obfuscation using `garble`. These scripts provide a complete build pipeline with backup management, cross-platform compatibility, and production-ready binary generation.

## Key Features

- **Automatic Tool Installation**: Checks and installs required tools (`goversioninfo`, `garble`)
- **User Choice for Obfuscation**: Interactive prompt to enable/disable code obfuscation
- **Timeout Handling**: 10-second timeout with default behavior (no obfuscation)
- **Cross-Platform**: Native implementations for Unix/Linux/macOS and Windows
- **Backup Management**: Automatic backup and archiving of previous builds
- **Process Management**: Stops running application instances before building
- **Asset Copying**: Automatically copies configuration and web assets

## Enhanced Build Process

### Tool Installation Phase

1. **Check goversioninfo**: Verifies if `goversioninfo` is installed
   - If not found: Automatically installs `github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest`
   - If found: Continues to next step

2. **Check garble**: Verifies if `garble` is installed
   - If not found: Automatically installs `mvdan.cc/garble@latest`
   - If found: Continues to next step

### User Choice Phase

3. **Obfuscation Prompt**: Interactive prompt with timeout
   ```
   Use garble build for obfuscation? (y/N, timeout 10s):
   ```
   - **Y/y**: Enables code obfuscation using `garble build`
   - **N/n or timeout**: Uses standard `go build`
   - **Default**: No obfuscation (safer for development)

### Standard Build Phase

4. **Version Info Generation**: Runs `goversioninfo -platform-specific`
5. **Binary Compilation**: Builds with appropriate command
   - With obfuscation: `garble build -ldflags="-s -w"`
   - Without obfuscation: `go build -ldflags="-s -w"`
6. **Asset Management**: Copies configuration files, web assets, and databases

## Usage Examples

### Unix/Linux/macOS

```bash
# Interactive build with user choice
./scripts/build.sh

# Example output:
#    (\_/)
/#    (o.o)   stackyard Builder by diameter-tscd
#   c(")(")
# ------------------------------------------------------------------------------
# [0/6] Checking required tools...
# + goversioninfo found
# + garble found
# Use garble build for obfuscation? (y/N, timeout 10s): y
# + Using garble build
# [1/6] Checking for running process...
# + App is not running.
# [2/6] Backing up old files...
# + Backup created at: dist/backups/20251220_235500
# [3/6] Archiving backup...
# + Backup archived: dist/backups/20251220_235500.zip
# [4/6] Building Go binary...
# + Build successful: dist/stackyard
# [5/6] Copying assets...
# + Copying web folder...
# + Copying config.yaml...
# SUCCESS! Build ready at: dist/
```

### Windows

```cmd
# Interactive build with user choice
scripts\build.bat

# Example output:
#    (\_/)
#    (o.o)   stackyard Builder by diameter-tscd
#   c(")(")
# ------------------------------------------------------------------------------
# [0/6] Checking required tools...
# + goversioninfo found
# + garble found
# Use garble build for obfuscation? (Y/N, default N, timeout 10s): Y
# + Using garble build
# [1/6] Checking for running process...
# + App is not running.
# [2/6] Backing up old files...
# + Backup created at: dist/backups/20251220_235500
# [3/6] Archiving backup...
# + Backup archived: dist/backups/20251220_235500.zip
# [4/6] Building Go binary...
# + Build successful: dist/stackyard.exe
# [5/6] Copying assets...
# + Copying web folder...
# SUCCESS! Build ready at: dist\
```

## Code Obfuscation with Garble

### What is Garble?

Garble is a Go code obfuscation tool that:
- **Obfuscates identifiers**: Renames functions, variables, and types
- **Removes debug info**: Strips file paths and source information
- **Maintains functionality**: Preserves program behavior
- **Increases binary size**: Obfuscated binaries are slightly larger

### When to Use Obfuscation

**Recommended for:**
- Production deployments
- Commercial applications
- Security-sensitive code
- Intellectual property protection

**Not recommended for:**
- Development builds
- Debugging scenarios
- Open source projects
- Performance-critical applications (minor overhead)

### Obfuscation Effects

```go
// Original code
package main

func calculateTotal(items []Item) int {
    total := 0
    for _, item := range items {
        total += item.price
    }
    return total
}

// Obfuscated result (example)
package main

func A(items []B) int {
    C := 0
    for _, D := range items {
        C += D.E
    }
    return C
}
```

## Configuration Options

### Build Script Variables

| Variable | Unix/Linux/macOS | Windows | Description |
|----------|------------------|---------|-------------|
| `DIST_DIR` | `dist` | `dist` | Output directory |
| `APP_NAME` | `stackyard` | `stackyard.exe` | Binary name |
| `MAIN_PATH` | `./cmd/app/main.go` | `./cmd/app/main.go` | Main Go file |

### Timeout Behavior

- **Timeout Duration**: 10 seconds
- **Default Choice**: N (no obfuscation)
- **Platform Differences**:
  - Unix: Uses `read -t` with signal handling
  - Windows: Uses `choice /T` command

## Error Handling

### Tool Installation Failures

```bash
# If goversioninfo installation fails
[0/6] Checking required tools...
! goversioninfo not found. Installing...
x Failed to install goversioninfo
# Script exits with error code
```

### Build Failures

```bash
# If Go compilation fails
[4/6] Building Go binary...
x Build FAILED! Exit code: 2
# Script exits with build error code
```

### Recovery Options

**Clean Rebuild:**
```bash
# Remove dist directory and rebuild
rm -rf dist/
./scripts/build.sh
```

**Skip Obfuscation:**
```bash
# Choose 'N' when prompted or wait for timeout
# Script will use standard go build
```

## Performance Comparison

### Build Times

| Build Type | Average Time | Binary Size | Notes |
|------------|--------------|-------------|-------|
| Standard Go | ~15-30s | ~15-25MB | Normal compilation |
| Garble Build | ~45-90s | ~18-28MB | Slower, larger binary |
| UPX Compressed | +10-20s | ~6-10MB | Additional compression |

### Runtime Performance

- **Standard Build**: Baseline performance
- **Garble Build**: ~1-5% slower due to obfuscated symbols
- **Memory Usage**: No significant difference

## Security Considerations

### Code Protection

**Obfuscation Benefits:**
- Makes reverse engineering more difficult
- Protects intellectual property
- Complicates debugging by attackers
- Reduces information leakage

**Limitations:**:
- Not encryption (can be deobfuscated with effort)
- Source code recovery is difficult but not impossible
- Performance debugging becomes harder

### Best Practices

1. **Development**: Use standard builds for easier debugging
2. **Staging**: Test with obfuscated builds before production
3. **Production**: Always use obfuscated builds for security
4. **CI/CD**: Automate obfuscation for consistent deployments

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: Build and Deploy

on:
  push:
    branches: [ main ]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3

    - name: Install Go tools
      run: |
        go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
        go install mvdan.cc/garble@latest

    - name: Build without obfuscation
      run: go build -ldflags="-s -w" -o app ./cmd/app/main.go

    - name: Build with obfuscation
      run: garble build -ldflags="-s -w" -o app-obfuscated ./cmd/app/main.go
```

### Docker Integration

```dockerfile
# Multi-stage build with obfuscation choice
FROM golang:1.21-alpine AS builder

# Install tools
RUN go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
RUN go install mvdan.cc/garble@latest

# Copy source
COPY . .

# Build with obfuscation (for production)
RUN goversioninfo -platform-specific
RUN garble build -ldflags="-s -w" -o main ./cmd/app

FROM alpine:latest
COPY --from=builder /app/main .
CMD ["./main"]
```

## Troubleshooting

### Common Issues

**"garble: command not found"**
- Ensure garble was installed successfully
- Check Go environment: `go env`
- Verify installation: `go list mvdan.cc/garble`

**"Build takes too long"**
- Garble builds are slower by design
- Consider using standard builds for development
- Use build caching in CI/CD

**"Obfuscated binary crashes"**
- Test obfuscated builds thoroughly before production
- Some reflection-based code may need adjustments
- Check for hardcoded function names

**"Timeout reached during prompt"**
- Script continues with default (no obfuscation)
- For automated builds, consider removing the prompt
- Use environment variables for CI/CD

### Debug Options

**Verbose Output:**
```bash
# Add to build script for debugging
set -x  # Unix
@echo on  # Windows
```

**Skip Prompt (Automated Builds):**
```bash
# Modify script to skip user interaction
USE_GARBLE=false  # Always use standard build
# Or
USE_GARBLE=true   # Always use obfuscation
```

## Migration Guide

### From Previous Version

1. **Backup existing scripts** (optional but recommended)
2. **Replace build scripts** with enhanced versions
3. **Test installation** of required tools
4. **Verify builds** work with both obfuscation options
5. **Update CI/CD** pipelines if needed

### Backward Compatibility

- **Existing builds**: Continue to work unchanged
- **Configuration**: Same environment variables and paths
- **Output format**: Identical directory structure
- **Dependencies**: Only adds optional tools

## Advanced Usage

### Custom Build Flags

Modify the build commands for additional flags:

```bash
# Unix/Linux/macOS
garble build -ldflags="-s -w -X main.version=1.2.3" -o "$DIST_DIR/$APP_NAME" "$MAIN_PATH"

# Windows
garble build -ldflags="-s -w -X main.version=1.2.3" -o "%DIST_DIR%\%APP_NAME%" %MAIN_PATH%
```

### Environment-Based Choice

For automated environments:

```bash
# Set environment variable
export USE_GARBLE=true

# Modify script to check environment
if [ "$USE_GARBLE" = "true" ]; then
    # Skip prompt, use garble
else
    # Show prompt
fi
```

### Multiple Build Targets

```bash
# Build both versions
./scripts/build.sh  # Interactive choice
USE_GARBLE=false ./scripts/build.sh  # Standard only
USE_GARBLE=true ./scripts/build.sh   # Obfuscated only
```

## Conclusion

The enhanced build scripts provide a robust, user-friendly solution for Go application building with optional code obfuscation. The interactive prompts ensure developers can make informed choices about security vs. development convenience, while the automatic tool installation removes setup barriers.

Key benefits:
- **Security**: Optional code obfuscation for production deployments
- **Usability**: Interactive prompts with sensible defaults
- **Automation**: CI/CD friendly with timeout handling
- **Reliability**: Comprehensive error handling and backup management
- **Cross-platform**: Native implementations for all major platforms
