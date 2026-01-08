# Onboarding Scripts Documentation

## Overview

The Onboarding Scripts provide an interactive setup wizard for quickly configuring and initializing the stackyard application. These cross-platform scripts guide new users through the essential configuration steps, ensuring a smooth first-time setup experience.

## Features

- **Interactive Configuration**: Step-by-step guided setup process
- **Cross-Platform Support**: Separate scripts for Unix/Linux/macOS and Windows
- **Comprehensive Setup**: Configures app settings, ports, services, and infrastructure
- **Security Awareness**: Highlights default credentials that must be changed
- **Automatic Validation**: Tests build and dependency setup
- **Backup Safety**: Creates backups before making configuration changes
- **User-Friendly**: Color-coded output with clear instructions and defaults

## Prerequisites

### Unix/Linux/macOS
- Bash shell
- `config.yaml` file in the project root
- Go compiler (optional, for build testing)

### Windows
- Windows Command Prompt or PowerShell
- `config.yaml` file in the project root
- PowerShell (recommended for YAML manipulation)

## Quick Start

### Unix/Linux/macOS
```bash
# Make executable (first time only)
chmod +x scripts/onboarding.sh

# Run the onboarding script
./scripts/onboarding.sh
```

### Windows
```cmd
# Run the onboarding script
scripts\onboarding.bat
```

## Configuration Workflow

The onboarding script guides you through these configuration categories:

### 1. Basic Application Configuration
- **Application Name**: Display name for your application
- **Version**: Application version number
- **Server Port**: HTTP API server port (default: 8080)
- **Monitoring Port**: Web monitoring interface port (default: 9090)

### 2. Environment Settings
- **Debug Mode**: Enable detailed logging (default: Yes)
- **TUI Mode**: Enable Terminal User Interface (default: Yes)
- **Quiet Startup**: Suppress console logs during boot (default: No)

### 3. Service Configuration
- **Monitoring Dashboard**: Enable web-based monitoring interface (default: Yes)
- **API Encryption**: Enable end-to-end API encryption (default: No)

### 4. Infrastructure Configuration
- **PostgreSQL**: Database setup (single/multi/none, default: single)
- **Redis**: Caching layer (default: No)
- **Kafka**: Message queue (default: No)
- **MinIO**: Object storage (default: No)

## Configuration Examples

### Development Setup
```
Application Name: My Dev App
Version: 1.0.0
Server Port: 8080
Monitoring Port: 9090
Debug Mode: Yes
TUI Mode: Yes
Quiet Startup: No
Monitoring: Yes
Encryption: No
PostgreSQL: single
Redis: No
Kafka: No
MinIO: No
```

### Production Setup
```
Application Name: My Production App
Version: 1.0.0
Server Port: 8080
Monitoring Port: 9090
Debug Mode: No
TUI Mode: No
Quiet Startup: Yes
Monitoring: Yes
Encryption: Yes
PostgreSQL: multi
Redis: Yes
Kafka: Yes
MinIO: Yes
```

## Security Warnings

The onboarding script prominently displays security warnings about:

### Default Credentials (MUST Change)
- **PostgreSQL Password**: `Mypostgres01`
- **Monitoring Password**: `admin`
- **MinIO Credentials**: `minioadmin/minioadmin`
- **API Secret Key**: `super-secret-key`

### Security Features Information
- **API Obfuscation**: Enabled by default (security through obscurity)
- **Encryption**: Optional AES-256-GCM encryption for API communications

### Production Readiness Checklist
- [ ] Change all default passwords
- [ ] Configure strong encryption keys (if encryption enabled)
- [ ] Set up proper SSL/TLS certificates
- [ ] Configure firewall rules
- [ ] Enable audit logging
- [ ] Set up monitoring alerts

## Generated Configuration

The script updates `config.yaml` with your selections. Example output:

```yaml
app:
  name: "My Fancy Go App"
  version: "1.0.0"
  debug: true
  enable_tui: true
  quiet_startup: false

server:
  port: "8080"

monitoring:
  port: "9090"
  enabled: true

services:
  service_a: true
  service_b: false
  service_encryption: false

postgres:
  enabled: true
  connections:
    - name: "primary"
      enabled: true
      host: "localhost"
      port: 5432
      user: "postgres"
      password: "Mypostgres01"
      dbname: "postgres"
      sslmode: "disable"

redis:
  enabled: false

kafka:
  enabled: false

monitoring:
  minio:
    enabled: false

encryption:
  enabled: false
```

## Next Steps After Onboarding

The script provides clear guidance for post-setup tasks:

### 1. Security Configuration
```bash
# Update passwords in config.yaml
nano config.yaml

# Set environment variables for sensitive data
export POSTGRES_PASSWORD="your-secure-password"
export MONITORING_PASSWORD="your-admin-password"
```

### 2. Infrastructure Setup
```bash
# Start PostgreSQL
docker run -d --name postgres \
  -e POSTGRES_PASSWORD=your-secure-password \
  -p 5432:5432 postgres:15

# Start Redis (if enabled)
docker run -d --name redis -p 6379:6379 redis:7

# Configure MinIO (if enabled)
docker run -d --name minio \
  -p 9000:9000 \
  -e MINIO_ACCESS_KEY=your-access-key \
  -e MINIO_SECRET_KEY=your-secret-key \
  minio/minio server /data
```

### 3. Build and Test
```bash
# Update dependencies
go mod tidy

# Build the application
./scripts/build.sh

# Test the application
go run cmd/app/main.go
```

### 4. Production Deployment
```bash
# Build Docker images
./scripts/docker_build.sh "myapp" "myregistry/myapp"

# Deploy with Docker Compose
docker-compose up -d

# Access monitoring dashboard
open http://localhost:9090
```

## Troubleshooting

### Common Issues

**"config.yaml not found"**
- Ensure you're running the script from the project root directory
- Check that `config.yaml` exists: `ls -la config.yaml`

**"Permission denied" (Unix/Linux/macOS)**
- Make the script executable: `chmod +x scripts/onboarding.sh`
- Check file permissions on the scripts directory

**"PowerShell not available" (Windows)**
- The script will fall back to manual configuration instructions
- Install PowerShell or manually edit `config.yaml`

**Configuration not applied**
- Check for syntax errors in `config.yaml`
- Restore from backup if needed: `cp config.yaml.backup config.yaml`

**Build test fails**
- Ensure Go is installed: `go version`
- Check dependencies: `go mod tidy`
- Verify configuration syntax

### Recovery Options

**Restore from Backup**
```bash
# Unix/Linux/macOS
cp config.yaml.backup config.yaml

# Windows
copy config.yaml.backup config.yaml
```

**Reset to Defaults**
```bash
# Remove current config and use template
rm config.yaml
cp config.template.yaml config.yaml
```

**Manual Configuration**
If the script fails, manually edit `config.yaml` with a text editor following the examples in this documentation.

## Advanced Usage

### Automated Setup (CI/CD)
```bash
# Non-interactive mode (future enhancement)
./scripts/onboarding.sh --non-interactive --config production.yaml

# Custom configuration file
./scripts/onboarding.sh --config my-config.yaml
```

### Custom Infrastructure
For complex setups, modify `config.yaml` after running the onboarding script:

```yaml
# Multi-tenant PostgreSQL setup
postgres:
  enabled: true
  connections:
    - name: "tenant_a"
      host: "db-tenant-a.company.com"
      sslmode: "require"
    - name: "tenant_b"
      host: "db-tenant-b.company.com"
      sslmode: "require"

# High availability Redis
redis:
  enabled: true
  cluster: true
  addresses:
    - "redis-1.company.com:6379"
    - "redis-2.company.com:6379"
    - "redis-3.company.com:6379"
```

## Script Architecture

### Unix/Linux/macOS Implementation (`onboarding.sh`)
- **Bash scripting** with color support
- **YAML manipulation** using `sed` for simple key-value updates
- **Error handling** with backup restoration on failure
- **Interactive prompts** with validation and defaults

### Windows Implementation (`onboarding.bat`)
- **Batch scripting** with ANSI color support
- **PowerShell integration** for complex YAML manipulation
- **Fallback handling** for systems without PowerShell
- **Cross-compatibility** with Windows CMD limitations

## Integration with Other Scripts

The onboarding script works seamlessly with other project scripts:

```bash
# Run onboarding first
./scripts/onboarding.sh

# Then build
./scripts/build.sh

# Or build Docker images
./scripts/docker_build.sh

# Change package name if needed
./scripts/change_package.sh "github.com/mycompany/myproject"
```

## Best Practices

### Development Environment
1. Run onboarding with development defaults
2. Keep debug mode enabled
3. Use TUI for better development experience
4. Enable monitoring for development insights

### Production Environment
1. Run onboarding with production settings
2. Disable debug mode and TUI
3. Enable quiet startup for cleaner logs
4. Configure all security settings
5. Set up proper infrastructure (SSL, firewalls, monitoring)

### Team Onboarding
1. Commit the configured `config.yaml` (without secrets)
2. Document environment variable requirements
3. Create setup documentation for new team members
4. Use version control for configuration templates

## Support and Resources

### Documentation Links
- [Configuration Guide](CONFIGURATION_GUIDE.md) - Complete config reference
- [Build Scripts](BUILD_SCRIPTS.md) - Application building
- [Docker Containerization](DOCKER_CONTAINERIZATION.md) - Container deployment
- [Integration Guide](INTEGRATION_GUIDE.md) - Infrastructure setup

### Community Resources
- GitHub Issues: Report bugs and request features
- Discussions: Share setup experiences and tips
- Wiki: Extended documentation and examples
