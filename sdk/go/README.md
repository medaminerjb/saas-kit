# SaaSKit Go SDK

The official Go client library for SaaSKit, providing strongly-typed access to the SaaSKit API with built-in retry controls and backoff algorithms.

## Installation

```bash
go get github.com/medaminerjb/saas-kit-go
```

## Quick Start

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
    // Create a new client
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
    fmt.Printf("Registered user: %s\n", registerResp.Email)

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

    // Update user metadata
    metadata, err := client.Metadata.UpdateUserMetadata(ctx, accessToken, &saaskit.UpdateUserMetadataRequest{
        MetadataPublic: map[string]interface{}{
            "theme":        "dark",
            "preferences": map[string]string{"language": "en"},
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Updated metadata: %+v\n", metadata)
}
```

## Configuration

The client can be configured with the following options:

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `BaseURL` | string | `http://localhost:8080` | Base URL of the SaaSKit API |
| `APIKey` | string | - | Optional API key for authentication |
| `HTTPClient` | `*http.Client` | nil (default client) | Custom HTTP client |
| `Timeout` | `time.Duration` | `30s` | HTTP request timeout |
| `MaxRetries` | int | `3` | Maximum number of retries for failed requests |
| `RetryDelay` | `time.Duration` | `1s` | Initial delay between retries |

## Retry Logic

The SDK automatically retries failed requests with exponential backoff:

- Retries are attempted for server errors (5xx status codes)
- Retry delay follows exponential backoff: `delay * 2^(attempt-1)`
- Maximum retry delay is capped at 30 seconds
- Client errors (4xx) are not retried

## Authentication

### Register

```go
resp, err := client.Auth.Register(ctx, &saaskit.RegisterRequest{
    Email:    "user@example.com",
    Password: "securepassword123",
})
```

### Login

```go
resp, err := client.Auth.Login(ctx, &saaskit.LoginRequest{
    Email:    "user@example.com",
    Password: "securepassword123",
})
accessToken := resp.AccessToken
refreshToken := resp.RefreshToken
```

### Refresh Token

```go
resp, err := client.Auth.Refresh(ctx, &saaskit.RefreshRequest{
    RefreshToken: refreshToken,
})
```

### Logout

```go
err := client.Auth.Logout(ctx, accessToken, &saaskit.LogoutRequest{
    RefreshToken: refreshToken,
})
```

### Forgot Password

```go
err := client.Auth.ForgotPassword(ctx, &saaskit.ForgotPasswordRequest{
    Email: "user@example.com",
})
```

### Reset Password

```go
err := client.Auth.ResetPassword(ctx, &saaskit.ResetPasswordRequest{
    Token:    "reset-token",
    Password: "newpassword123",
})
```

### Verify Email

```go
err := client.Auth.VerifyEmail(ctx, &saaskit.VerifyEmailRequest{
    Token: "verification-token",
})
```

## User Management

### Get Current User

```go
user, err := client.Users.GetMe(ctx, accessToken)
```

### Update Current User

```go
user, err := client.Users.UpdateMe(ctx, accessToken, &saaskit.UpdateUserRequest{
    FirstName: "John",
    LastName:  "Doe",
})
```

### List Sessions

```go
sessions, err := client.Users.ListSessions(ctx, accessToken)
```

### Revoke Session

```go
err := client.Users.RevokeSession(ctx, accessToken, sessionID)
```

## Tenant Management

### Create Tenant

```go
tenant, err := client.Tenants.Create(ctx, accessToken, &saaskit.CreateTenantRequest{
    Name: "My Organization",
    Slug: "my-org",
})
```

### List Tenants

```go
tenants, err := client.Tenants.List(ctx, accessToken)
for _, t := range tenants {
    fmt.Printf("%s (role: %s)\n", t.Tenant.Name, t.Role)
}
```

### Get Tenant

```go
tenant, err := client.Tenants.Get(ctx, accessToken, tenantID)
```

### Update Tenant

```go
tenant, err := client.Tenants.Update(ctx, accessToken, tenantID, &saaskit.UpdateTenantRequest{
    Name: "Updated Name",
})
```

### Switch Active Tenant

```go
err := client.Tenants.Switch(ctx, accessToken, &saaskit.SwitchTenantRequest{
    TenantID: tenantID,
})
```

### Accept Invitation

```go
err := client.Tenants.AcceptInvitation(ctx, accessToken, &saaskit.AcceptInvitationRequest{
    Token: "invitation-token",
})
```

### List Members

```go
members, err := client.Tenants.ListMembers(ctx, accessToken, tenantID)
```

### Invite Member

```go
member, err := client.Tenants.InviteMember(ctx, accessToken, tenantID, &saaskit.InviteMemberRequest{
    Email: "newuser@example.com",
    Role:  "member",
})
```

### Remove Member

```go
err := client.Tenants.RemoveMember(ctx, accessToken, tenantID, userID)
```

## Metadata Operations

### Get User Metadata

```go
metadata, err := client.Metadata.GetUserMetadata(ctx, accessToken)
```

### Update User Metadata

```go
metadata, err := client.Metadata.UpdateUserMetadata(ctx, accessToken, &saaskit.UpdateUserMetadataRequest{
    MetadataPublic: map[string]interface{}{
        "theme":        "dark",
        "preferences": map[string]string{"language": "en"},
    },
    MetadataPrivate: map[string]interface{}{
        "internal_flags": []string{"beta_user"},
    },
})
```

### Get Tenant Metadata

```go
metadata, err := client.Metadata.GetTenantMetadata(ctx, accessToken, tenantID)
```

### Update Tenant Metadata

```go
metadata, err := client.Metadata.UpdateTenantMetadata(ctx, accessToken, tenantID, &saaskit.UpdateTenantMetadataRequest{
    Metadata: map[string]interface{}{
        "billing_id": "cus_1234567890",
        "locale":     "en-US",
    },
})
```

## Error Handling

All SDK methods return errors that should be handled appropriately:

```go
user, err := client.Users.GetMe(ctx, accessToken)
if err != nil {
    // Handle error
    log.Printf("Failed to get user: %v", err)
    return
}
```

## Context Support

All SDK methods accept a `context.Context` for cancellation and timeouts:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

user, err := client.Users.GetMe(ctx, accessToken)
```

## License

Apache License 2.0
