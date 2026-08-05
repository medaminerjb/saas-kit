# SaaSKit JavaScript SDK

The official JavaScript/TypeScript client library for SaaSKit, providing type-safe access to the SaaSKit API with built-in retry controls and backoff algorithms.

## Installation

```bash
npm install @saaskit/js
# or
yarn add @saaskit/js
# or
pnpm add @saaskit/js
```

## Quick Start

```typescript
import SaaSKitClient from '@saaskit/js';

// Create a new client
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
console.log('Registered user:', registerResp.email);

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

// Update user metadata
const metadata = await client.metadata.updateUserMetadata(accessToken, {
  metadata_public: {
    theme: 'dark',
    preferences: { language: 'en' },
  },
});
console.log('Updated metadata:', metadata);
```

## Configuration

The client can be configured with the following options:

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `baseURL` | string | `http://localhost:8080` | Base URL of the SaaSKit API |
| `apiKey` | string | - | Optional API key for authentication |
| `timeout` | number | `30000` | HTTP request timeout in milliseconds |
| `maxRetries` | number | `3` | Maximum number of retries for failed requests |
| `retryDelay` | number | `1000` | Initial delay between retries in milliseconds |

## Retry Logic

The SDK automatically retries failed requests with exponential backoff:

- Retries are attempted for server errors (5xx status codes)
- Retry delay follows exponential backoff: `delay * 2^(attempt-1)`
- Maximum retry delay is capped at 30 seconds
- Client errors (4xx) are not retried

## Authentication

### Register

```typescript
const resp = await client.auth.register({
  email: 'user@example.com',
  password: 'securepassword123',
});
```

### Login

```typescript
const resp = await client.auth.login({
  email: 'user@example.com',
  password: 'securepassword123',
});
const accessToken = resp.access_token;
const refreshToken = resp.refresh_token;
```

### Refresh Token

```typescript
const resp = await client.auth.refresh({
  refresh_token: refreshToken,
});
```

### Logout

```typescript
await client.auth.logout(accessToken, {
  refresh_token: refreshToken,
});
```

### Forgot Password

```typescript
await client.auth.forgotPassword({
  email: 'user@example.com',
});
```

### Reset Password

```typescript
await client.auth.resetPassword({
  token: 'reset-token',
  password: 'newpassword123',
});
```

### Verify Email

```typescript
await client.auth.verifyEmail({
  token: 'verification-token',
});
```

## User Management

### Get Current User

```typescript
const user = await client.users.getMe(accessToken);
```

### Update Current User

```typescript
const user = await client.users.updateMe(accessToken, {
  first_name: 'John',
  last_name: 'Doe',
});
```

### List Sessions

```typescript
const sessions = await client.users.listSessions(accessToken);
```

### Revoke Session

```typescript
await client.users.revokeSession(accessToken, sessionId);
```

## Tenant Management

### Create Tenant

```typescript
const tenant = await client.tenants.create(accessToken, {
  name: 'My Organization',
  slug: 'my-org',
});
```

### List Tenants

```typescript
const tenants = await client.tenants.list(accessToken);
for (const t of tenants) {
  console.log(`${t.tenant.name} (role: ${t.role})`);
}
```

### Get Tenant

```typescript
const tenant = await client.tenants.get(accessToken, tenantId);
```

### Update Tenant

```typescript
const tenant = await client.tenants.update(accessToken, tenantId, {
  name: 'Updated Name',
});
```

### Switch Active Tenant

```typescript
await client.tenants.switch(accessToken, {
  tenant_id: tenantId,
});
```

### Accept Invitation

```typescript
await client.tenants.acceptInvitation(accessToken, {
  token: 'invitation-token',
});
```

### List Members

```typescript
const members = await client.tenants.listMembers(accessToken, tenantId);
```

### Invite Member

```typescript
const member = await client.tenants.inviteMember(accessToken, tenantId, {
  email: 'newuser@example.com',
  role: 'member',
});
```

### Remove Member

```typescript
await client.tenants.removeMember(accessToken, tenantId, userId);
```

## Metadata Operations

### Get User Metadata

```typescript
const metadata = await client.metadata.getUserMetadata(accessToken);
```

### Update User Metadata

```typescript
const metadata = await client.metadata.updateUserMetadata(accessToken, {
  metadata_public: {
    theme: 'dark',
    preferences: { language: 'en' },
  },
  metadata_private: {
    internal_flags: ['beta_user'],
  },
});
```

### Get Tenant Metadata

```typescript
const metadata = await client.metadata.getTenantMetadata(accessToken, tenantId);
```

### Update Tenant Metadata

```typescript
const metadata = await client.metadata.updateTenantMetadata(accessToken, tenantId, {
  metadata: {
    billing_id: 'cus_1234567890',
    locale: 'en-US',
  },
});
```

## Error Handling

All SDK methods throw `SaaSKitError` on failure:

```typescript
try {
  const user = await client.users.getMe(accessToken);
} catch (error) {
  if (error instanceof SaaSKitError) {
    console.error(`Failed to get user: ${error.message}`);
    console.error(`Status: ${error.status}`);
    console.error(`Response:`, error.response);
  }
}
```

## TypeScript Support

The SDK is written in TypeScript and provides full type definitions:

```typescript
import SaaSKitClient, { User, Tenant, LoginResponse } from '@saaskit/js';

const client = new SaaSKitClient();
const loginResp: LoginResponse = await client.auth.login({ /* ... */ });
const user: User = await client.users.getMe(loginResp.access_token);
```

## License

Apache License 2.0
