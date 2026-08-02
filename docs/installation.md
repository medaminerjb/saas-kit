# SaaSKit Installation Guide

This guide details the requirements and steps for installing, setting up, and running SaaSKit v0.1.0 in both local development and production environments.

---

## 1. Prerequisites

Before installing SaaSKit, ensure you have the following software installed:

* **Go**: Version 1.26.1 or later.
* **PostgreSQL**: Version 16.x or later.
* **Docker & Docker Compose**: (Optional but highly recommended for local database setup and production packaging).
* **Goose**: CLI tool (optional, for running database migrations manually).

---

## 2. Local Development Setup

To run SaaSKit locally for development:

### Step 1: Clone the Repository
```bash
git clone https://github.com/saaskit/saaskit.git
cd saaskit
```

### Step 2: Start Infrastructure Dependencies
Use the provided `docker-compose.yml` to start a PostgreSQL 16 instance:
```bash
docker compose up -d postgres
```

### Step 3: Run Database Migrations
SaaSKit uses `goose` to manage schemas. You can apply the migrations using the Makefile helper:
```bash
# If goose is installed:
make migrate-up

# Alternatively, using docker exec to migrate through goose inside the container if configured,
# or running it using SaaSKit's programmatic migration system.
```

### Step 4: Configure Environment Variables
Copy the sample environment file and adjust configuration parameters:
```bash
cp .env.example .env
```
Refer to the [Configuration Guide](configuration.md) for details on all configuration keys.

### Step 5: Start the Server
Run the application locally in development mode:
```bash
go run cmd/saaskit/main.go
```
The server will start on port `8080` by default. You can verify it is healthy by curling:
```bash
curl http://localhost:8080/health
```

---

## 3. Running Tests

To verify that SaaSKit is working as expected on your system:

### Run Unit Tests
```bash
go test -v -short ./...
```

### Run Integration & E2E Tests
Make sure the local PostgreSQL database is running and configured correctly in your environment variables, then run:
```bash
SAASKIT_DATABASE_HOST=localhost \
SAASKIT_DATABASE_PORT=5432 \
SAASKIT_DATABASE_USER=postgres \
SAASKIT_DATABASE_PASSWORD=postgres \
SAASKIT_DATABASE_NAME=saaskit_test \
SAASKIT_DATABASE_SSLMODE=disable \
go test -v ./...
```

---

## 4. Production Deployment

SaaSKit is built to be cloud-native and runs seamlessly as a Docker container.

### Step 1: Build the Docker Image
A multi-stage, hardened production `Dockerfile` is provided in the repository:
```bash
docker build -t saaskit:latest -f docker/Dockerfile .
```

### Step 2: Run the Container
Deploy the image to your container engine (e.g., Kubernetes, ECS, or Docker Compose) ensuring that all required production environment variables are set:
```bash
docker run -d \
  -p 8080:8080 \
  --name saaskit-app \
  -e SAASKIT_ENV=production \
  -e SAASKIT_DATABASE_HOST=prod-db-host \
  -e SAASKIT_DATABASE_PORT=5432 \
  -e SAASKIT_DATABASE_USER=saaskit_admin \
  -e SAASKIT_DATABASE_PASSWORD=secure-db-password \
  -e SAASKIT_DATABASE_NAME=saaskit \
  -e SAASKIT_SERVER_SECRET=production-long-secure-server-secret-key \
  -e SAASKIT_ENCRYPTION_MASTER_KEY=6368616e676520746869732070617373776f726420746f206120736563726574 \
  saaskit:latest
```

> [!IMPORTANT]
> In production, you **MUST** configure a secure, unique `SAASKIT_SERVER_SECRET` and `SAASKIT_ENCRYPTION_MASTER_KEY` (32-byte hex string) to protect user credentials, session cookies, and database state.
