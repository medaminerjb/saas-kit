# SaaSKit SDK Integration Guide

This guide provides detailed integration examples and best practices for using the SaaSKit SDKs in production applications.

## Table of Contents

- [Go SDK Integration](#go-sdk-integration)
- [JavaScript SDK Integration](#javascript-sdk-integration)
- [Authentication Patterns](#authentication-patterns)
- [Error Handling](#error-handling)
- [Testing](#testing)
- [Best Practices](#best-practices)

## Go SDK Integration

### Setting Up the Client

Create a singleton client instance in your application:

```go
package main

import (
    "time"
    "github.com/medaminerjb/saas-kit-go"
)

var saaskitClient *saaskit.Client

func InitSaaSKit(baseURL string) error {
    var err error
    saaskitClient, err = saaskit.NewClient(&saaskit.Config{
        BaseURL:    baseURL,
        Timeout:    30 * time.Second,
        MaxRetries: 3,
        RetryDelay: 1 * time.Second,
    })
    return err
}

func GetSaaSKitClient() *saaskit.Client {
    return saaskitClient
}
```

### Middleware for Authentication

Create middleware to inject access tokens into requests:

```go
package middleware

import (
    "net/http"
    "strings"
)

type AuthMiddleware struct {
    tokenProvider func() string
}

func NewAuthMiddleware(tokenProvider func() string) *AuthMiddleware {
    return &AuthMiddleware{tokenProvider: tokenProvider}
}

func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := m.tokenProvider()
        if token != "" {
            r.Header.Set("Authorization", "Bearer "+token)
        }
        next.ServeHTTP(w, r)
    })
}
```

### Token Management

Implement token refresh logic:

```go
package auth

import (
    "context"
    "sync"
    "time"
    
    "github.com/medaminerjb/saas-kit-go"
)

type TokenManager struct {
    client       *saaskit.Client
    accessToken  string
    refreshToken string
    mu           sync.RWMutex
    expiresAt    time.Time
}

func NewTokenManager(client *saaskit.Client) *TokenManager {
    return &TokenManager{client: client}
}

func (tm *TokenManager) Login(ctx context.Context, email, password string) error {
    resp, err := tm.client.Auth.Login(ctx, &saaskit.LoginRequest{
        Email:    email,
        Password: password,
    })
    if err != nil {
        return err
    }

    tm.mu.Lock()
    defer tm.mu.Unlock()
    
    tm.accessToken = resp.AccessToken
    tm.refreshToken = resp.RefreshToken
    tm.expiresAt = time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
    
    return nil
}

func (tm *TokenManager) GetAccessToken() string {
    tm.mu.RLock()
    defer tm.mu.RUnlock()
    
    // Refresh if expiring soon
    if time.Until(tm.expiresAt) < 5*time.Minute {
        go tm.refresh()
    }
    
    return tm.accessToken
}

func (tm *TokenManager) refresh() error {
    ctx := context.Background()
    resp, err := tm.client.Auth.Refresh(ctx, &saaskit.RefreshRequest{
        RefreshToken: tm.refreshToken,
    })
    if err != nil {
        return err
    }

    tm.mu.Lock()
    defer tm.mu.Unlock()
    
    tm.accessToken = resp.AccessToken
    tm.refreshToken = resp.RefreshToken
    tm.expiresAt = time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
    
    return nil
}
```

### Tenant Context Management

Manage tenant context for multi-tenant applications:

```go
package tenant

import (
    "context"
    "github.com/google/uuid"
)

type contextKey string

const TenantIDKey contextKey = "tenant_id"

func WithTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
    return context.WithValue(ctx, TenantIDKey, tenantID)
}

func GetTenantID(ctx context.Context) (uuid.UUID, bool) {
    val := ctx.Value(TenantIDKey)
    if val == nil {
        return uuid.Nil, false
    }
    tenantID, ok := val.(uuid.UUID)
    return tenantID, ok
}
```

## JavaScript SDK Integration

### React Integration

Create a React context for authentication:

```typescript
// src/context/AuthContext.tsx
import React, { createContext, useContext, useState, useEffect } from 'react';
import SaaSKitClient, { LoginResponse, User } from '@saaskit/js';

interface AuthContextType {
  client: SaaSKitClient;
  user: User | null;
  accessToken: string | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  loading: boolean;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [client] = useState(() => new SaaSKitClient({
    baseURL: process.env.VITE_SAASKIT_URL || 'http://localhost:8080',
  }));
  const [user, setUser] = useState<User | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const login = async (email: string, password: string) => {
    const resp = await client.auth.login({ email, password });
    setAccessToken(resp.access_token);
    localStorage.setItem('refresh_token', resp.refresh_token);
    
    const currentUser = await client.users.getMe(resp.access_token);
    setUser(currentUser);
  };

  const logout = async () => {
    if (accessToken) {
      const refreshToken = localStorage.getItem('refresh_token');
      if (refreshToken) {
        await client.auth.logout(accessToken, { refresh_token: refreshToken });
      }
    }
    setAccessToken(null);
    setUser(null);
    localStorage.removeItem('refresh_token');
  };

  useEffect(() => {
    // Check for existing session
    const initAuth = async () => {
      const token = localStorage.getItem('access_token');
      if (token) {
        try {
          const currentUser = await client.users.getMe(token);
          setUser(currentUser);
          setAccessToken(token);
        } catch (error) {
          localStorage.removeItem('access_token');
        }
      }
      setLoading(false);
    };
    
    initAuth();
  }, [client]);

  return (
    <AuthContext.Provider value={{ client, user, accessToken, login, logout, loading }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
```

### Vue.js Integration

Create a Vue plugin for SaaSKit:

```typescript
// src/plugins/saaskit.ts
import { App } from 'vue';
import SaaSKitClient from '@saaskit/js';

export default {
  install(app: App) {
    const client = new SaaSKitClient({
      baseURL: import.meta.env.VITE_SAASKIT_URL || 'http://localhost:8080',
    });
    
    app.provide('saaskit', client);
    app.config.globalProperties.$saaskit = client;
  }
};
```

Use in components:

```vue
<script setup lang="ts">
import { inject } from 'vue';
import SaaSKitClient from '@saaskit/js';

const saaskit = inject<SaaSKitClient>('saaskit');

const login = async () => {
  const resp = await saaskit!.auth.login({
    email: 'user@example.com',
    password: 'password',
  });
  // Store token...
};
</script>
```

### Next.js Integration

Create API route proxy for SaaSKit:

```typescript
// pages/api/auth/login.ts
import type { NextApiRequest, NextApiResponse } from 'next';
import SaaSKitClient from '@saaskit/js';

const client = new SaaSKitClient({
  baseURL: process.env.SAASKIT_URL || 'http://localhost:8080',
});

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse
) {
  if (req.method !== 'POST') {
    return res.status(405).json({ error: 'Method not allowed' });
  }

  try {
    const { email, password } = req.body;
    const resp = await client.auth.login({ email, password });
    
    // Set HTTP-only cookie
    res.setHeader('Set-Cookie', [
      `access_token=${resp.access_token}; HttpOnly; Path=/; Max-Age=${resp.expires_in}`,
      `refresh_token=${resp.refresh_token}; HttpOnly; Path=/; Max-Age=604800`,
    ]);
    
    res.status(200).json({ success: true });
  } catch (error) {
    res.status(401).json({ error: 'Authentication failed' });
  }
}
```

## Authentication Patterns

### API Key Authentication

Use API keys for service-to-service communication:

```go
// Go SDK with API key
client, err := saaskit.NewClient(&saaskit.Config{
    BaseURL: "http://localhost:8080",
    APIKey:  "sk_live_abc123...",
})
```

```typescript
// JavaScript SDK with API key
const client = new SaaSKitClient({
  baseURL: 'http://localhost:8080',
  apiKey: 'sk_live_abc123...',
});
```

### Token Refresh Pattern

Implement automatic token refresh:

```go
func (tm *TokenManager) DoRequest(ctx context.Context, fn func(token string) error) error {
    token := tm.GetAccessToken()
    err := fn(token)
    
    if err != nil && strings.Contains(err.Error(), "401") {
        // Token expired, refresh and retry
        if refreshErr := tm.refresh(); refreshErr != nil {
            return refreshErr
        }
        return fn(tm.GetAccessToken())
    }
    
    return err
}
```

## Error Handling

### Go SDK Error Handling

```go
func handleUserError(err error) {
    if err != nil {
        if strings.Contains(err.Error(), "401") {
            // Handle unauthorized
            log.Println("Authentication required")
        } else if strings.Contains(err.Error(), "403") {
            // Handle forbidden
            log.Println("Insufficient permissions")
        } else if strings.Contains(err.Error(), "404") {
            // Handle not found
            log.Println("Resource not found")
        } else {
            // Handle other errors
            log.Printf("Error: %v", err)
        }
    }
}
```

### JavaScript SDK Error Handling

```typescript
try {
  const user = await client.users.getMe(accessToken);
} catch (error) {
  if (error instanceof SaaSKitError) {
    switch (error.status) {
      case 401:
        // Redirect to login
        window.location.href = '/login';
        break;
      case 403:
        // Show permission denied message
        alert('Insufficient permissions');
        break;
      case 404:
        // Show not found message
        alert('Resource not found');
        break;
      default:
        // Show generic error
        alert(`Error: ${error.message}`);
    }
  }
}
```

## Testing

### Go SDK Testing

Use a test client with mock server:

```go
func TestUserManagement(t *testing.T) {
    // Create test server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Mock responses
    }))
    defer server.Close()
    
    client, _ := saaskit.NewClient(&saaskit.Config{
        BaseURL: server.URL,
    })
    
    // Test operations
    user, err := client.Users.GetMe(context.Background(), "test-token")
    assert.NoError(t, err)
    assert.NotNil(t, user)
}
```

### JavaScript SDK Testing

Use MSW for mocking HTTP requests:

```typescript
import { rest } from 'msw';
import { setupServer } from 'msw/node';

const server = setupServer(
  rest.post('http://localhost:8080/api/v1/auth/login', (req, res, ctx) => {
    return res(
      ctx.status(200),
      ctx.json({
        access_token: 'test-token',
        refresh_token: 'test-refresh',
        token_type: 'Bearer',
        expires_in: 3600,
      })
    );
  })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

test('login', async () => {
  const client = new SaaSKitClient({ baseURL: 'http://localhost:8080' });
  const resp = await client.auth.login({
    email: 'test@example.com',
    password: 'password',
  });
  
  expect(resp.access_token).toBe('test-token');
});
```

## Best Practices

### Security

1. **Never store tokens in localStorage** for sensitive applications. Use HTTP-only cookies or secure storage.
2. **Use HTTPS** in production for all API communications.
3. **Validate tokens** on the server side for sensitive operations.
4. **Implement rate limiting** to prevent abuse.
5. **Use short-lived access tokens** with refresh tokens.

### Performance

1. **Reuse client instances** instead of creating new ones for each request.
2. **Implement connection pooling** for Go SDK HTTP clients.
3. **Cache frequently accessed data** like user profiles.
4. **Use context timeouts** to prevent hanging requests.

### Error Handling

1. **Always handle errors** from SDK methods.
2. **Implement retry logic** for transient failures (SDK has built-in retry).
3. **Log errors** for debugging and monitoring.
4. **Provide user-friendly error messages** in your application.

### Multi-Tenancy

1. **Always include tenant context** in requests when using tenant-scoped operations.
2. **Validate tenant membership** before allowing access to tenant resources.
3. **Use tenant-specific API keys** for service-to-service communication.
4. **Implement tenant isolation** in your application logic.

## Support

For more information, visit the [SaaSKit repository](https://github.com/medaminerjb/saas-kit) or open an issue on GitHub.
