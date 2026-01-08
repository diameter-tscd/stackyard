# Package Name Change Scripts

## Overview

The Package Name Change Scripts provide automated tools for renaming the Go module package name across the entire codebase. These scripts are essential when refactoring, renaming, or migrating a Go project to a new module path.

## Features

- **Cross-Platform Support**: Separate implementations for Unix/Linux/macOS and Windows
- **Comprehensive Updates**: Updates both the module declaration and all import paths
- **Safety Mechanisms**: Backup creation and error validation
- **Pure Native Tools**: No external dependencies required
- **Recursive Processing**: Handles all `.go` files in the project directory tree

## Scripts Overview

### change_package.sh (Unix/Linux/macOS)

A Bash script that uses standard Unix utilities for text processing and file manipulation.

**Prerequisites:**
- Bash shell
- `sed` (stream editor)
- `find` (file search utility)
- `awk` (text processing)
- `grep` (pattern matching)

### change_package.bat (Windows)

A Windows batch script that uses pure CMD commands for maximum compatibility.

**Prerequisites:**
- Windows Command Prompt
- No external tools required (uses built-in CMD features)

## Usage

### Basic Usage

**Unix/Linux/macOS:**
```bash
chmod +x scripts/change_package.sh
./scripts/change_package.sh "github.com/new-org/new-project"
```

**Windows:**
```cmd
scripts\change_package.bat github.com/new-org/new-project
```

### Examples

```bash
# Change to a GitHub repository
./scripts/change_package.sh "github.com/mycompany/myproject"

# Change to a local/private module
./scripts/change_package.sh "mycompany.com/internal/project"

# Change to a generic domain
scripts\change_package.bat "example.com/my-app"

# Change to a subdirectory module
./scripts/change_package.sh "github.com/user/project/v2"
```

## How It Works

The scripts perform the following operations in sequence:

### 1. Validation Phase

- **Input Validation**: Ensures a new module name is provided as argument
- **Module Discovery**: Reads the current module name from `go.mod`
- **Error Checking**: Validates that `go.mod` exists and contains a valid module declaration

### 2. Update Phase

- **Module Declaration**: Updates the `module` line in `go.mod`
- **Import Path Scanning**: Recursively finds all `.go` files in the project
- **Import Path Replacement**: Updates all import statements containing the old module name

### 3. Completion Phase

- **Success Reporting**: Displays completion message with statistics
- **Backup Notification**: Reminds user about backup files (Unix/Linux/macOS)

## Technical Implementation

### Unix/Linux/macOS Implementation

The Bash script uses:

```bash
# Read current module name
CURRENT_MODULE=$(grep '^module ' go.mod | awk '{print $2}')

# Update go.mod with backup
sed -i.bak "s|^module $CURRENT_MODULE|module $NEW_MODULE|" go.mod

# Update all .go files with backup
find . -name "*.go" -type f -exec sed -i.bak "s|$CURRENT_MODULE|$NEW_MODULE|g" {} +
```

### Windows Implementation

The batch script uses CMD's built-in string replacement:

```batch
REM Enable delayed expansion for variable manipulation
setlocal enabledelayedexpansion

REM Extract current module name
for /f "tokens=2" %%i in ('findstr "^module " go.mod') do set CURRENT_MODULE=%%i

REM Replace in go.mod using temporary file
(for /f "delims=" %%i in (go.mod) do (
  set "line=%%i"
  set "line=!line:%search%=%replace%!"
  echo !line!
)) > "%tempfile%"

REM Replace in all .go files
for /r %%f in (*.go) do (
  REM Similar replacement logic for each file
)
```

## Safety Features

### Backup Creation

**Unix/Linux/macOS:**
- Creates `.bak` backup files for all modified files
- Preserves original content in case of errors
- Easy cleanup after verification

**Windows:**
- No automatic backups (CMD limitations)
- Recommends committing changes before running
- Files are overwritten directly

### Error Handling

- **Argument Validation**: Stops if no new module name provided
- **File Existence**: Checks for `go.mod` presence
- **Module Detection**: Validates current module name extraction
- **Operation Verification**: Checks success of file operations

### Recovery Options

If something goes wrong:

1. **Restore from backups** (Unix/Linux/macOS):
   ```bash
   find . -name "*.bak" -exec bash -c 'mv "$1" "${1%.bak}"' _ {} \;
   ```

2. **Git revert** (if changes were committed):
   ```bash
   git checkout HEAD~1 go.mod
   git checkout HEAD~1 -- "*.go"
   ```

## Best Practices

### Before Running

1. **Commit Current Changes**:
   ```bash
   git add .
   git commit -m "Before package name change"
   ```

2. **Test Compilation**:
   ```bash
   go build ./...
   go test ./...
   ```

3. **Backup Important Files** (additional safety):
   ```bash
   cp go.mod go.mod.backup
   ```

### After Running

1. **Verify Changes**:
   ```bash
   grep -r "old-module-name" . --include="*.go"
   ```

2. **Clean Up Backups** (Unix/Linux/macOS):
   ```bash
   find . -name "*.bak" -delete
   ```

3. **Update Dependencies**:
   ```bash
   go mod tidy
   ```

4. **Test Again**:
   ```bash
   go build ./...
   go test ./...
   ```

5. **Update IDE/Editor**:
   - Restart your IDE
   - Clear any cached module information

## Common Use Cases

### Repository Rename

When moving a project to a new GitHub organization:

```bash
# Old: github.com/old-org/project
# New: github.com/new-org/project
./scripts/change_package.sh "github.com/new-org/project"
```

### Versioning

When creating a new major version:

```bash
# Old: github.com/user/project
# New: github.com/user/project/v2
./scripts/change_package.sh "github.com/user/project/v2"
```

### Internal Migration

When moving from public to private repositories:

```bash
# Old: github.com/user/project
# New: company.com/internal/project
./scripts/change_package.sh "company.com/internal/project"
```

## Troubleshooting

### "Module declaration not found"

**Cause**: `go.mod` file missing or malformed
**Solution**: Ensure you're in the correct project directory and `go.mod` exists

### "Permission denied"

**Cause**: File permissions or write access issues
**Solution**: Check file permissions and run with appropriate privileges

### "sed: command not found" (Unix/Linux/macOS)

**Cause**: `sed` not installed
**Solution**: Install sed utility (usually pre-installed on most systems)

### "FINDSTR is not recognized" (Windows)

**Cause**: Running in PowerShell instead of CMD
**Solution**: Use `cmd.exe` or modify script for PowerShell compatibility

### Import Paths Not Updated

**Cause**: Custom import aliases or complex import statements
**Solution**: Manually review and update any remaining imports

### IDE Still Shows Old Imports

**Cause**: IDE caching module information
**Solution**: Restart IDE and clear module cache

## Performance Considerations

- **Large Codebases**: The script processes all `.go` files recursively
- **File Count**: Performance scales with the number of Go files
- **Backup Overhead**: Unix script creates backups (extra I/O)
- **Memory Usage**: Minimal memory footprint for both implementations

## Integration with CI/CD

### GitHub Actions Example

```yaml
- name: Change Package Name
  run: |
    chmod +x scripts/change_package.sh
    ./scripts/change_package.sh "github.com/${{ github.repository }}"

- name: Verify Changes
  run: |
    go mod tidy
    go build ./...
```

### Jenkins Pipeline Example

```groovy
stage('Package Rename') {
    steps {
        sh 'chmod +x scripts/change_package.sh'
        sh "./scripts/change_package.sh ${NEW_MODULE_NAME}"
    }
}

stage('Verification') {
    steps {
        sh 'go mod tidy'
        sh 'go build ./...'
    }
}
```

## Security Considerations

- **Input Validation**: Scripts validate module name format
- **Path Safety**: Operations confined to project directory
- **No Network Access**: Scripts work entirely offline
- **No External Dependencies**: Reduces supply chain risks

## Future Enhancements

Potential improvements for future versions:

- **Dry Run Mode**: Preview changes without applying them
- **Selective Updates**: Update only specific directories
- **Import Alias Handling**: Better support for custom import aliases
- **Validation Checks**: Verify new module name format
- **Progress Indicators**: Show progress for large codebases
- **Undo Functionality**: Automated rollback capability

## Conclusion

The Package Name Change Scripts provide a reliable, cross-platform solution for Go module renaming. They handle the complexity of updating both module declarations and import paths while providing safety mechanisms to prevent data loss.

The scripts follow the principle of "convention over configuration" - they work automatically with standard Go project structures without requiring additional setup or configuration files.
