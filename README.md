
<div align="center">
  <img src=".assets/stackyrd-banner.png" alt="stackyrd" style="width: 100%; max-width: 700px;"/>
</div>
<div align="center">
  <img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License"/>
  <img src="https://img.shields.io/badge/go-1.21%2B-00ADD8.svg" alt="Go Version"/>
  <img src="https://img.shields.io/badge/build-passing-brightgreen.svg" alt="Build Status"/>
  <img src="https://img.shields.io/badge/github-diameter--tscd/stackyrd-181717.svg" alt="GitHub Repo"/>
</div>
<br>

stackyrd is a production-ready application framework featuring modular architecture, comprehensive monitoring, real-time dashboards, and extensive infrastructure integrations. Built for scalability and ease of deployment.

## Quick Start

### Prerequisites
- Go 1.21+

### Installation & Run

```bash
# Clone the repository
git clone https://github.com/diameter-tscd/stackyrd.git
cd stackyrd

# Install dependencies
go mod download

# Run the application
go run cmd/app/main.go

# To build the application
go run scripts/build/build.go

```

**First Access:**
1. Open `http://localhost:9090` (monitoring dashboard)
2. Login with password: `admin`
3. **Important**: Change the default password immediately!

## Screenshots

### CLI UI
![Console](.assets/Recording_Console.gif)

## Key Features

- **Modular Services**: Enable/disable services via configuration
- **Monitoring Dashboard**: Real-time metrics, logs, and system monitoring
- **Terminal UI**: Interactive boot sequence and live CLI dashboard
- **Infrastructure Support**: Redis, PostgreSQL (multi-tenant), Kafka, MinIO
- **Security**: API encryption, authentication, and access controls
- **Build Tools**: Automated build scripts with backup and archiving

## Documentation

**[Full Documentation](docs_wiki/)** - Comprehensive guides and references

## License

Apache License Version 2.0: [LICENSE](LICENSE)


**Powered by diameter-tscd**
