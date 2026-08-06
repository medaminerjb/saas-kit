# SaaSKit SDK Quickstart Guide

This guide will help you get started with the SaaSKit SDKs for Go and JavaScript.

## Prerequisites

- A running SaaSKit instance (see [README](../README.md) for setup instructions)
- Go 1.22+ (for Go SDK)
- Node.js 18+ (for JavaScript SDK)

## Go SDK Quickstart

### Installation

```bash
go get github.com/medaminerjb/saas-kit-go
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/medaminerjb/saas-kit-go"
)

func main() {
    // Initialize the client
    client, err := saaskit.NewClient(&saaskit.Config{
        BaseURL:    "http://localhost:8080",
        Timeout:    30 * time.Second,
        MaxRetries: 3,
        RetryDelay: 1 * time.Second,
    })
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Register a new user
    registerResp, err := client.Auth.Register(ctx, &saaskit.RegisterRequest{
        Email:    "user@example.com",
        Password: "securepassword123",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Registered: %s\n", registerResp.Email)

    // Login
    loginResp, err := client.Auth.Login(ctx, &saaskit.LoginRequest{
        Email:    "user@example.com",
        Password: "securepassword123",
    })
    if err != nil {
        log.Fatal(err)
    }
    accessToken := loginResp.AccessToken

    // Get current user
    user, err := client.Users.GetMe(ctx, accessToken)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Current user: %s\n", user.Email)

    // Create a tenant
    tenant, err := client.Tenants.Create(ctx, accessToken, &saaskit.CreateTenantRequest{
        Name: "My Organization",
        Slug: "my-org",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created tenant: %s\n", tenant.Name)
}
```

### Using with Context

All SDK methods accept a `context.Context` for cancellation and timeouts:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

user, err := client.Users.GetMe(ctx, accessToken)
```

## JavaScript SDK Quickstart

### Installation

```bash
npm install @saaskit/js
# or
yarn add @saaskit/js
# or
pnpm add @saaskit/js
```

### Basic Usage

```typescript
import SaaSKitClient from '@saaskit/js';

// Initialize the client
const client = new SaaSKitClient({
  baseURL: 'http://localhost:8080',
  timeout: 30000,
  maxRetries: 3,
  retryDelay: 1000,
});

// Register a new user
const registerResp = await client.auth.register({
  email: 'user@example.com',
  password: 'securepassword123',
});
console.log('Registered:', registerResp.email);

// Login
const loginResp = await client.auth.login({
  email: 'user@example.com',
  password: 'securepassword123',
});
const accessToken = loginResp.access_token;

// Get current user
const user = await client.users.getMe(accessToken);
console.log('Current user:', user.email);

// Create a tenant
const tenant = await client.tenants.create(accessToken, {
  name: 'My Organization',
  slug: 'my-org',
});
console.log('Created tenant:', tenant.name);
```

### Error Handling

```typescript
try {
  const user = await client.users.getMe(accessToken);
} catch (error) {
  if (error instanceof SaaSKitError) {
    console.error(`Failed: ${error.message}`);
    console.error(`Status: ${error.status}`);
  }
}
```

## CLI Quickstart

The SaaSKit CLI provides commands for managing users, tenants, and API keys.

### Installation

The CLI is built with the main SaaSKit project. Build it using:

```bash
go build -o saaskit ./cmd/cli
```

### User Management

Create a new user:

```bash
./saaskit user create \
  --email user@example.com \
  --password securepassword123
```

### Tenant Management

Create a new tenant:

```bash
./saaskit tenant create \
  --name "My Organization" \
  --slug my-org \
  --owner-email user@example.com
```

List tenants for a user:

```bash
./saaskit tenant list --email user@example.com
```

Switch active tenant:

```bash
./saaskit tenant switch \
  --email user@example.com \
  --tenant-id <tenant-id>
```

### API Key Management

Create an API key:

```bash
./saaskit apikey create \
  --email user@example.com \
  --tenant-id <tenant-id> \
  --name "Production Key" \
  --scopes tenant.read,tenant.write
```

List API keys for a tenant:

```bash
./saaskit apikey list --tenant-id <tenant-id>
```

Revoke an API key:

```bash
./saaskit apikey revoke --key-id <key-id>
```

## Configuration

### Environment Variables

The CLI can be configured using environment variables:

```bash
export DATABASE_URL="postgres://saaskit:saaskit@localhost:5432/saaskit?sslmode=disable"
export SERVER_SECRET="your-secret-key"
export ENCRYPTION_MASTER_KEY="your-32-byte-hex-key"
```

### SDK Configuration

Both SDKs support the following configuration options:

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `baseURL` | string | `http://localhost:8080` | Base URL of the SaaSKit API |
| `timeout` | number/duration | `30000`/`30s` | HTTP request timeout |
| `maxRetries` | number/int | `3` | Maximum number of retries |
| `retryDelay` | number/duration | `1000`/`1s` | Initial retry delay |

## Next Steps

- Read the full [Go SDK documentation](../sdk/go/README.md)
- Read the full [JavaScript SDK documentation](../sdk/js/README.md)
- Explore the [API endpoints documentation](../README.md#api-endpoints)
- Check out the [integration guides](./sdk-integration.md)

## Support

For more information, visit the [SaaSKit repository](https://github.com/medaminerjb/saas-kit) or open an issue on GitHub.
