# Welcome to the Project Documentation

This wiki serves as the central knowledge base for the project, covering architecture, feature implementations, and integration guides.

## Core Documentation

### Architecture & Design

*   **[Async Infrastructure Implementation](ASYNC_INFRASTRUCTURE.md)**
    *   Complete async infrastructure system ensuring non-blocking operations.
    *   Worker pools, goroutines, and channels for concurrent processing.
    *   Generic AsyncResult types with timeout and error handling.
    *   Performance benefits and best practices for async operations.

*   **[Configuration Guide](CONFIGURATION_GUIDE.md)**
    *   Complete reference for `config.yaml` configuration.
    *   All available options with explanations and examples.
    *   Multiple PostgreSQL connections setup guide with web monitoring interface.
    *   Dynamic connection switching and health monitoring.

*   **[Architecture Diagrams](ARCHITECTURE_DIAGRAMS.md)**
    *   Visual guide to the system's request/response flow.
    *   Package organization and dependency graphs.
    *   Sequence diagrams for validation and error handling.

*   **[API Response Structure](API_RESPONSE_STRUCTURE.md)**
    *   Standard format for all API responses (`success`, `data`, `meta`).
    *   Built-in helper functions for success and error responses.
    *   Pagination and validation standards.

*   **[Request Response Structure](REQUEST_RESPONSE_STRUCTURE.md)**
    *   Detailed overview of the Echo service structure.
    *   Request validation patterns and custom validators.
    *   Dependencies and best practices for creating new endpoints.

### Security & Privacy

*   **[API Obfuscation](API_OBFUSCATION.md)**
    *   Mechanism for obscuring JSON data in transit using Base64.
    *   Configuration guide for enabling/disabling obfuscation.
    *   Frontend and Backend implementation details.

*   **[API Request/Response Encryption](ENCRYPTION_API.md)**
    *   End-to-end encryption for all API communications using AES-256-GCM.
    *   Automatic middleware for transparent encryption/decryption.
    *   Key management, rotation, and secure storage.
    *   Client implementation guides for JavaScript and Python.
    *   Configuration, security best practices, and troubleshooting.

### User Interface

*   **[TUI Implementation](TUI_IMPLEMENTATION.md)**
    *   Documentation for the Terminal User Interface (Bubble Tea).
    *   **Boot Sequence**: Visual feedback during service initialization.
    *   **Live Logs**: Real-time log display with scrolling, filtering, and management controls.
    *   **Enhanced Controls**: Keyboard shortcuts for scrolling, filtering, auto-scroll toggle, and log clearing.
    *   **Reusable Dialog System**: Template-based dialog components for easy reuse.
    *   **Unlimited Log Storage**: Removed 1000 log limit for unlimited storage.
    *   **Default Auto-scroll**: Auto-scroll enabled by default on application startup.

### Build & Deployment

*   **[Build Scripts](BUILD_SCRIPTS.md)**
    *   Enhanced build scripts with code obfuscation and cross-platform support.
    *   **Customizable Parameter Parsing System**: Dynamic flag system for command-line configuration.
    *   **Tool Installation**: Automatic installation of required Go tools (`goversioninfo`, `garble`).
    *   **Code Obfuscation**: Optional `garble` build for production security.
    *   **Cross-Platform Support**: Native implementations for Unix/Linux/macOS and Windows.

### Real-time Features

*   **[Live Event Streaming](EVENT_STREAMING.md)**
    *   Server-Sent Events (SSE) implementation with multiple concurrent streams.
    *   Real-time push notifications to connected clients without polling.
    *   Event broadcasting to specific streams or all streams simultaneously.
    *   Stream management with start, stop, pause, and resume operations.
    *   Client management with automatic subscription/unsubscription.

### Integration & Infrastructure

*   **[MongoDB Integration](MONGODB_INTEGRATION.md)**
    *   Complete MongoDB integration guide with multiple database support.
    *   Web monitoring interface for MongoDB databases and collections.
    *   Manual query execution and real-time database statistics.
    *   Multi-tenant MongoDB operations with connection switching.
    *   CRUD operations, aggregation pipelines, and best practices.

*   **[Grafana Integration](GRAFANA_INTEGRATION.md)**
    *   Complete Grafana API integration for dashboard and data source management.
    *   Programmatic dashboard creation, updates, and deletion.
    *   Data source configuration and annotation support.
    *   Async operations with retry logic and health monitoring.

*   **[Docker Containerization](DOCKER_CONTAINERIZATION.md)**
    *   Multi-stage Dockerfile for development, testing, and production.
    *   Docker Compose integration with infrastructure services.
    *   CI/CD integration, security best practices, and troubleshooting.

*   **[Build Scripts](BUILD_SCRIPTS.md)**
    *   Automated build process with backup and archiving.
    *   Cross-platform scripts for Unix/Linux/macOS and Windows.
    *   Backup management, troubleshooting, and CI/CD integration.

*   **[Package Name Change Scripts](CHANGE_PACKAGE_SCRIPTS.md)**
    *   Automated tools for renaming Go module package names.
    *   Cross-platform support for comprehensive codebase refactoring.
    *   Safety mechanisms, backup creation, and error handling.

*   **[Onboarding Scripts](ONBOARDING_SCRIPTS.md)**
    *   Interactive setup wizard for first-time configuration.
    *   Cross-platform scripts for Unix/Linux/macOS and Windows.
    *   Guided configuration of app settings, services, and infrastructure.
    *   Security warnings and best practices for production setup.

*   **[Service Implementation Guide](SERVICE_IMPLEMENTATION.md)**
    *   Creating and implementing new services.
    *   Service interface and registration.
    *   Dynamic configuration setup.

*   **[Integration Guide](INTEGRATION_GUIDE.md)**
    *   **Redis**: Configuration and usage of the Redis manager.
    *   **Postgres**: Database connection and Helper methods.
    *   Multiple PostgreSQL connections with dynamic switching in monitoring UI.
    *   **MongoDB**: NoSQL database integration with multi-connection support.
    *   **Kafka**: Message producing and configuration.
    *   **MinIO**: Object storage integration for file uploads.
    *   **Cron Jobs**: Dynamic job scheduling and management.

### Examples & Samples

*   **[Service F (Multi-Tenant Orders)](../internal/services/modules/service_f.go)**
    *   Complete example of using multiple PostgreSQL connections.
    *   Demonstrates tenant-based database isolation.
    *   Shows dynamic connection selection in API endpoints.

---

## Getting Started

If you are new to the project, we recommend starting with the **[Integration Guide](INTEGRATION_GUIDE.md)** to understand the available infrastructure components, followed by the **[API Response Structure](API_RESPONSE_STRUCTURE.md)** to learn how to build consistent API endpoints.
