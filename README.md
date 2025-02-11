# HTTP Server

An extensible template for building HTTP servers.

## Table of Contents

- [Features](#features)
- [Technologies Used](#technologies-used)
- [Modules](#modules)
- [Prerequisites](#prerequisites)

## Features

- RESTful API
- Web user interface
- Basic authentication mechanism
- Graceful shutdown capabilities
<!-- - gRPC server and client -->
- Read configuration from YAML file
  - Specify server port
  - Enable modules by listing them in the `modules` field ([learn more about it](#modules))
- Docker containerization
- Docker Compose for container orchestration
- Pre-populated PostgreSQL database integration
<!-- - Redis integration for request counting -->
- Nginx as a reverse proxy for:
  - TLS termination
  - Load balancing
  - Redirecting HTTP traffic to HTTPS
    <!-- - Configurable request and response timeouts -->
    <!-- - Rate limiting to control request traffic -->
- Middleware for:
  - Authentication
  - Request logging
  - Media type enforcement for data-driven requests
  <!-- - Unit and integration testing capabilities -->
- Continuous integration with GitHub Actions
  <!-- - CI/CD pipeline with Jenkins -->
  <!-- - Export Prometheus metrics -->
  <!-- - Metrics visualization with Grafana -->
- Health check endpoint at `/healthz`
- Development shell scripts

## Technologies Used

- Go
- Docker & Docker Compose
- Nginx
- PostgreSQL
- GitHub Actions

## Modules

This application features a modular design that allows you to toggle specific functionalities through a configuration file. The key components include:

- **Authentication:**
  - Module name: `auth`
- **Database:**
  - Module name: `database`
  - Utilizes a persistent database when enabled.j
  - Falls back to in-memory storage when disabled.
- **Web Interface:**
  - Module name: `webui`
  - Served at the root URL (`/`) when enabled.
  - Redirects to the API when disabled.
  <!-- - **gRPC:**
  - Module name: `grpc`
  - ... -->
  <!-- - **Dashboard:**
  - Module name: `dashboard`
  - ... -->

Modules can be enabled by listing their names in the `modules` field of the `config.yaml` file.

## Prerequisites

- Go (version 1.22 or later)
- Required environmental variables:
  - `AUTH_USERNAME`
  - `AUTH_PASSWORD`
  - `DB_USERNAME`
  - `DB_PASSWORD`
- Valid TLS certificates located in the `/certs` directory (default filenames: `nginx-selfsigned.crt`, `nginx-selfsigned.key`). You can generate these certificates using the following command:

```sh
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout nginx-selfsigned.key -out nginx-selfsigned.crt
```

---

[**Architecture**](/docs/architecture.md)
