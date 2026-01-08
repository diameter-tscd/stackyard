# Docker Containerization Guide

## Overview

This project includes comprehensive Docker containerization with multi-stage builds for development, testing, and production environments. The Docker setup provides consistent deployment across different environments while optimizing for security, performance, and maintainability.

## Dockerfile Architecture

The `Dockerfile` implements a multi-stage build strategy with optimized stages for minimal image sizes:

### 1. Builder Stage (Optimized)

```dockerfile
FROM golang:1.25.5-alpine3.23 AS builder

WORKDIR /app

# Install build dependencies and UPX for compression
RUN apk add --no-cache upx

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-w -s" \
    -trimpath \
    -o main ./cmd/app

# Compress binary with UPX (ultra-brute for maximum compression)
RUN upx --ultra-brute main
```

**Optimizations Applied:**
- **UPX Compression**: Reduces binary size by 50-70% using ultra-brute compression
- **Build Flags**: `-ldflags="-w -s"` removes debugging information and symbol table
- **Trimpath**: Removes file system paths from the compiled executable
- **Static Linking**: `CGO_ENABLED=0` ensures fully static binaries

**Purpose**: Compiles the Go application into a static binary
- Uses Alpine Linux for smaller base image
- Downloads dependencies separately for better layer caching
- Produces a statically linked binary with `CGO_ENABLED=0`
- Targets Linux platform for container compatibility

### 2. Test Stage

```dockerfile
FROM builder AS test

# Run tests
RUN go test ./...
```

**Purpose**: Executes the test suite in an isolated environment
- Inherits all source code and dependencies from builder stage
- Runs `go test ./...` to execute all test packages
- Can be targeted separately for CI/CD testing pipelines

### 3. Development Stage

```dockerfile
FROM golang:1.25.5-alpine3.23 AS dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN go build -o main ./cmd/app

# Configure for Docker environment
ENV APP_QUIET_STARTUP=false
ENV APP_ENABLE_TUI=false

# Expose ports for main API server and monitoring server
EXPOSE 8080 9090

# Run the application
CMD ["./main"]
```

**Purpose**: Provides a development environment with hot-reload capabilities
- Includes full Go toolchain for development tools
- Mounts source code for live development
- Exposes both main API (8080) and monitoring (9090) ports
- Automatically disables TUI and quiet startup for containerized environment

### 4. Production Stage

```dockerfile
FROM alpine:latest AS prod

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Configure for Docker environment
ENV APP_QUIET_STARTUP=false
ENV APP_ENABLE_TUI=false

# Expose ports for main API server and monitoring server
EXPOSE 8080 9090

# Run the application
CMD ["./main"]
```

**Purpose**: Creates a minimal production image
- Uses Alpine Linux for security and small size
- Includes only the compiled binary and runtime dependencies
- No source code or build tools included
- Optimized for production deployment
- Automatically disables TUI and quiet startup for containerized environment

## Building Docker Images

The project includes automated build scripts for building Docker images across different environments.

### Using Build Scripts

#### Unix/Linux/macOS

```bash
# Make script executable (first time only)
chmod +x scripts/docker_build.sh

# Build with default settings
./scripts/docker_build.sh

# Build with custom app name and image name
./scripts/docker_build.sh "my-app" "myregistry/myapp"
```

#### Windows

```cmd
# Build with default settings
scripts\docker_build.bat

# Build with custom app name and image name
scripts\docker_build.bat "my-app" "myregistry/myapp"
```

**Script Parameters:**
- `APP_NAME`: Application name (default: "stackyard")
- `IMAGE_NAME`: Docker image name (default: "myapp")
- `TARGET`: Build target - "all", "test", "dev", or "prod" (default: "all")

**What the scripts do:**
1. **Test Stage**: Builds and runs tests to ensure code quality (for "test" and "all" targets)
2. **Development Stage**: Builds development image with full Go toolchain (for "dev" and "all" targets)
3. **Production Stage**: Builds optimized production image (for "prod" and "all" targets)
4. **Cleanup**: Removes dangling Docker images to save space

**Target Options:**
- `all`: Build all stages (test, dev, prod) - default behavior
- `test`: Build and run tests only (2 steps)
- `dev`: Build development image only (1 step)
- `prod`: Build production image only (1 step) - Alpine (~50MB) with full monitoring
- `prod-slim`: Build slim production image (1 step) - Ubuntu (~30-40MB) with full monitoring
- `prod-minimal`: Build minimal production image (1 step) - BusyBox (~10-20MB) with full monitoring
- `ultra-prod`: Build ultra-minimal production image only (1 step) - smallest size using Distroless (~15-30MB, no monitoring)
- `ultra-all`: Build all ultra-minimal stages (ultra-test, ultra-dev, ultra-prod)
- `ultra-dev`: Build ultra-minimal development image (Distroless) - runs pre-built binary only
- `ultra-test`: Build ultra-minimal test image (Distroless) - runs pre-built binary only

**Usage Examples:**
```bash
# Build everything (default)
./scripts/docker_build.sh

# Build only production image (fastest)
./scripts/docker_build.sh "myapp" "myregistry/myapp" "prod"

# Build slim production image (~30-40MB, more secure)
./scripts/docker_build.sh "myapp" "myregistry/myapp" "prod-slim"

# Build minimal production image (~10-20MB)
./scripts/docker_build.sh "myapp" "myregistry/myapp" "prod-minimal"

# Build ultra-minimal production image (smallest, no monitoring)
./scripts/docker_build.sh "myapp" "myregistry/myapp" "ultra-prod"

# Build everything with ultra-prod for production
./scripts/docker_build.sh "myapp" "myregistry/myapp" "ultra-all"
```

### Manual Docker Commands

If you prefer to build manually:

#### Development Build

```bash
# Build development image
docker build --target dev -t myapp:dev .

# Run development container
docker run -p 8080:8080 -p 9090:9090 myapp:dev
```

#### Testing Build

```bash
# Build and run tests
docker build --target test -t myapp:test .

# Run tests only (will exit after tests complete)
docker run myapp:test
```

#### Production Build

```bash
# Build production image
docker build --target prod -t myapp:latest .

# Run production container
docker run -p 8080:8080 -p 9090:9090 myapp:latest
```

## Configuration in Containers

### Environment Variables

The application supports configuration via environment variables that override `config.yaml`:

```bash
# Run with custom configuration
docker run \
  -e SERVER_PORT=3000 \
  -e MONITORING_PORT=4000 \
  -e REDIS_ENABLED=true \
  -e REDIS_HOST=redis-server \
  -p 3000:3000 \
  -p 4000:4000 \
  myapp:latest
```

### Volume Mounts

For development with live reloading:

```bash
# Mount config file
docker run \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -p 8080:8080 \
  -p 9090:9090 \
  myapp:dev
```

For production with external config:

```bash
# Mount external config
docker run \
  -v /path/to/config.yaml:/root/config.yaml \
  -p 8080:8080 \
  -p 9090:9090 \
  myapp:latest
```

## Networking

### Port Configuration

The application exposes two main ports:

- **8080**: Main API server port (configurable via `server.port`)
- **9090**: Monitoring web interface port (configurable via `monitoring.port`)

### Port Mapping Examples

```bash
# Default port mapping
docker run -p 8080:8080 -p 9090:9090 myapp:latest

# Custom host ports
docker run -p 3000:8080 -p 4000:9090 myapp:latest

# Bind to specific interface
docker run -p 127.0.0.1:8080:8080 myapp:latest
```

## Docker Compose Integration

### Basic docker-compose.yml

```yaml
version: '3.8'

services:
  app:
    build:
      context: .
      target: prod
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - SERVER_PORT=8080
      - MONITORING_PORT=9090
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=myapp
      - POSTGRES_USER=myapp
      - POSTGRES_PASSWORD=password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

### Development docker-compose.yml

```yaml
version: '3.8'

services:
  app:
    build:
      context: .
      target: dev
    ports:
      - "8080:8080"
      - "9090:9090"
    volumes:
      - .:/app
      - /app/main
    environment:
      - APP_ENV=development
      - APP_DEBUG=true
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=myapp_dev
      - POSTGRES_USER=myapp
      - POSTGRES_PASSWORD=password
    volumes:
      - postgres_dev_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_dev_data:
```

## Multi-Stage Build Benefits

### Security
- **Minimal Attack Surface**: Production images contain only the runtime binary
- **No Source Code**: Source code is not included in production images
- **Alpine Linux**: Uses secure, minimal base images with regular security updates

### Performance
- **Small Image Size**: Multi-stage builds reduce final image size significantly
- **Fast Startup**: Optimized binary compilation with static linking
- **Layer Caching**: Dependencies are cached in separate layers for faster rebuilds

### Maintainability
- **Clear Separation**: Each stage has a specific purpose and can be built independently
- **Targeted Builds**: Can build specific stages for different environments
- **Easy Debugging**: Development stage includes full toolchain for troubleshooting

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build and Deploy

on:
  push:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Build test image
      run: docker build --target test -t myapp:test .
    - name: Run tests
      run: docker run myapp:test

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Build production image
      run: docker build --target prod -t myapp:latest .
    - name: Push to registry
      run: |
        echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin
        docker tag myapp:latest myregistry/myapp:latest
        docker push myregistry/myapp:latest
```

### Jenkins Pipeline Example

```groovy
pipeline {
    agent any

    stages {
        stage('Test') {
            steps {
                sh 'docker build --target test -t myapp:test .'
                sh 'docker run myapp:test'
            }
        }

        stage('Build') {
            steps {
                sh 'docker build --target prod -t myapp:latest .'
            }
        }

        stage('Deploy') {
            steps {
                sh 'docker-compose up -d'
            }
        }
    }
}
```

## Troubleshooting

### Common Issues

**"exec ./main: no such file or directory"**
- Ensure `CGO_ENABLED=0` in builder stage for static linking
- Check that the binary was built correctly in the builder stage

**"connection refused" to database**
- Ensure database containers are started before the app
- Use `depends_on` in docker-compose for proper startup order

**Large image size**
- Verify that production stage only copies the binary from builder
- Use `.dockerignore` to exclude unnecessary files

**Permission denied on volume mounts**
- Ensure proper file permissions on mounted directories
- Check user permissions in the container

### Debug Commands

```bash
# Check running containers
docker ps

# View container logs
docker logs <container_id>

# Execute shell in running container (not available in Distroless)
docker exec -it <container_id> /bin/sh

# Inspect image layers
docker history myapp:latest

# Check image size
docker images myapp

# Verify which base image was used
docker inspect myapp:ultra | grep -A 5 "RepoTags"
```

### Ultra-Prod Issues

**"Still shows Alpine image for ultra-prod"**
- **Cause**: Docker may be showing cached images or intermediate layers
- **Solution**: Check `docker images` for the specific tag, or run `docker build --no-cache --target ultra-prod -t myapp:ultra .`

**"Ultra-prod image not much smaller"**
- **Cause**: Binary size may still be large despite UPX compression
- **Solution**: Check binary size with `ls -lh main` in builder stage, ensure UPX is working

**"Cannot run ultra-prod container"**
- **Cause**: Distroless images have no shell, debugging is limited
- **Solution**: Use Alpine-based prod image for debugging, switch to ultra-prod for production

## Best Practices

### Security
1. **Use Official Images**: Base images from trusted sources only
2. **Regular Updates**: Keep base images and dependencies updated
3. **Minimal Images**: Use Alpine variants for smaller attack surface
4. **No Secrets in Images**: Use environment variables or secret management

### Performance
1. **Multi-Stage Builds**: Separate build and runtime stages
2. **Layer Optimization**: Order commands to maximize layer caching
3. **Minimal Base Images**: Use Alpine Linux for production
4. **Static Binaries**: Compile with `CGO_ENABLED=0` for portability

### Development Workflow
1. **Volume Mounts**: Mount source code for live development
2. **Hot Reload**: Use development stage with full Go toolchain
3. **Debug Tools**: Include debugging tools in development images
4. **Consistent Environments**: Use same base images across team

## Migration Guide

### From Single-Stage to Multi-Stage

**Before (single-stage):**
```dockerfile
FROM golang:1.21-alpine

WORKDIR /app
COPY . .
RUN go build -o main ./cmd/app

CMD ["./main"]
```

**After (multi-stage):**
```dockerfile
FROM golang:1.25.5-alpine3.23 AS builder
# ... build stage

FROM alpine:latest AS prod
# ... production stage
```

### Benefits of Migration
- **50-70% smaller images**: Remove build dependencies from final image
- **Better security**: No source code or build tools in production
- **Faster deployments**: Smaller images transfer and start faster
- **Flexible builds**: Can build different stages for different purposes

## Conclusion

The Docker containerization setup provides a robust, secure, and efficient way to deploy the Go application across different environments. The multi-stage build approach ensures optimal image sizes, security, and performance while maintaining flexibility for development and testing workflows.
