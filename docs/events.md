# SaaSKit Event Reference

This document describes all event types emitted by SaaSKit for webhook subscriptions and event-driven integrations.

## Event Format

All events follow this structure:

```json
{
  "type": "event.type",
  "tenant_id": "uuid (optional)",
  "actor_id": "uuid (optional)",
  "target_id": "uuid (optional)",
  "payload": { ... },
  "timestamp": "ISO-8601 timestamp"
}
```

## User Events

### `user.created`

Emitted when a new user account is created.

**Payload:**
```json
{
  "user_id": "uuid",
  "email": "string"
}
```

### `user.updated`

Emitted when a user's profile is updated.

**Payload:**
```json
{
  "user_id": "uuid",
  "changes": {
    "field": "value"
  }
}
```

### `user.deleted`

Emitted when a user account is deleted.

**Payload:**
```json
{
  "user_id": "uuid"
}
```

### `user.disabled`

Emitted when a user account is disabled (e.g., due to lockout).

**Payload:**
```json
{
  "user_id": "uuid",
  "reason": "string"
}
```

### `user.password_reset`

Emitted when a user successfully resets their password.

**Payload:**
```json
{
  "user_id": "uuid"
}
```

## Tenant Events

### `tenant.created`

Emitted when a new tenant (organization) is created.

**Payload:**
```json
{
  "tenant_id": "uuid",
  "name": "string",
  "slug": "string",
  "created_by": "uuid"
}
```

### `tenant.updated`

Emitted when tenant settings are updated.

**Payload:**
```json
{
  "tenant_id": "uuid",
  "changes": {
    "field": "value"
  }
}
```

### `tenant.deleted`

Emitted when a tenant is deleted.

**Payload:**
```json
{
  "tenant_id": "uuid"
}
```

### `member.invited`

Emitted when a user is invited to join a tenant.

**Payload:**
```json
{
  "tenant_id": "uuid",
  "email": "string",
  "role": "string",
  "invited_by": "uuid"
}
```

### `member.joined`

Emitted when a user accepts an invitation and joins a tenant.

**Payload:**
```json
{
  "tenant_id": "uuid",
  "user_id": "uuid",
  "role": "string"
}
```

### `member.removed`

Emitted when a member is removed from a tenant or leaves.

**Payload:**
```json
{
  "tenant_id": "uuid",
  "user_id": "uuid",
  "removed_by": "uuid (optional)"
}
```

### `member.updated`

Emitted when a member's role is updated.

**Payload:**
```json
{
  "tenant_id": "uuid",
  "user_id": "uuid",
  "old_role": "string",
  "new_role": "string",
  "updated_by": "uuid"
}
```

## API Key Events

### `api_key.created`

Emitted when a new API key is created.

**Payload:**
```json
{
  "api_key_id": "uuid",
  "tenant_id": "uuid",
  "name": "string",
  "type": "test|live",
  "scopes": ["scope1", "scope2"],
  "created_by": "uuid"
}
```

### `api_key.revoked`

Emitted when an API key is revoked.

**Payload:**
```json
{
  "api_key_id": "uuid",
  "tenant_id": "uuid",
  "revoked_by": "uuid"
}
```

### `api_key.deleted`

Emitted when an API key is permanently deleted.

**Payload:**
```json
{
  "api_key_id": "uuid",
  "tenant_id": "uuid"
}
```

## Session Events

### `session.created`

Emitted when a new user session is created (login).

**Payload:**
```json
{
  "session_id": "uuid",
  "user_id": "uuid",
  "tenant_id": "uuid (optional)",
  "ip_address": "string",
  "user_agent": "string"
}
```

### `session.revoked`

Emitted when a session is revoked (logout or admin action).

**Payload:**
```json
{
  "session_id": "uuid",
  "user_id": "uuid",
  "revoked_by": "uuid (optional)"
}
```

## Webhook Signature Verification

When a webhook subscription has a secret configured, SaaSKit signs the payload using HMAC-SHA256.

**Signature Header:** `X-SaaSKit-Signature`

**Format:** `sha256=<hex_signature>`

**Verification (Go example):**
```go
func verifySignature(payload []byte, secret string, signature string) bool {
    expectedPrefix := "sha256="
    if !strings.HasPrefix(signature, expectedPrefix) {
        return false
    }
    
    sig := signature[len(expectedPrefix):]
    h := hmac.New(sha256.New, []byte(secret))
    h.Write(payload)
    expected := hex.EncodeToString(h.Sum(nil))
    
    return hmac.Equal([]byte(sig), []byte(expected))
}
```

## Best Practices

1. **Idempotency:** Webhook endpoints should be idempotent. The same event may be delivered multiple times.
2. **Response Time:** Respond quickly (within 5 seconds). Use background processing for heavy operations.
3. **Status Codes:** Return `2xx` for successful delivery. Any other status code will trigger a retry.
4. **Ordering:** Events are delivered in order per subscription, but not guaranteed across subscriptions.
5. **Security:** Always verify webhook signatures if a secret is configured.
6. **Retries:** Failed deliveries are retried up to 3 times with exponential backoff.

## Event Delivery Guarantees

- **At-least-once:** Events may be delivered multiple times
- **Best-effort:** Events are delivered in the background; temporary failures are retried
- **No ordering:** Events are not guaranteed to be delivered in order across different subscriptions
- **TTL:** Failed delivery records are retained for 30 days
