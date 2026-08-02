# SaaSKit API Reference

This document provides a comprehensive API reference for SaaSKit v0.1.0's identity endpoints. All request and response bodies are formatted as JSON.

---

## 1. Authentication Endpoints

### Register User
Create a new user account.

* **URL**: `/api/v1/auth/register`
* **Method**: `POST`
* **Content-Type**: `application/json`

#### Request Body
```json
{
  "email": "user@example.com",
  "password": "Password123!",
  "name": "Jane Doe"
}
```

#### Response (201 Created)
```json
{
  "user": {
    "id": "cd8a405c-e168-4ac7-80f8-6a8da3fd6a42",
    "email": "user@example.com",
    "name": "Jane Doe",
    "status": "pending_verification",
    "email_verified": false,
    "created_at": "2026-08-02T18:46:57.819Z",
    "updated_at": "2026-08-02T18:46:57.819Z"
  },
  "tokens": {
    "access_token": "eyJhbGciOiJSUzI1NiIs...",
    "refresh_token": "AqsclfMk7mvGq2lSTBVs...",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

---

### Login User
Authenticate an existing user and retrieve session tokens.

* **URL**: `/api/v1/auth/login`
* **Method**: `POST`
* **Content-Type**: `application/json`

#### Request Body
```json
{
  "email": "user@example.com",
  "password": "Password123!"
}
```

#### Response (200 OK)
```json
{
  "user": {
    "id": "cd8a405c-e168-4ac7-80f8-6a8da3fd6a42",
    "email": "user@example.com",
    "name": "Jane Doe",
    "status": "pending_verification",
    "email_verified": false,
    "created_at": "2026-08-02T18:46:57.819Z",
    "updated_at": "2026-08-02T18:47:12.192Z"
  },
  "tokens": {
    "access_token": "eyJhbGciOiJSUzI1NiIs...",
    "refresh_token": "AqsclfMk7mvGq2lSTBVs...",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

#### Error Responses
* **401 Unauthorized**: Password incorrect.
* **403 Forbidden**: Account disabled or locked due to too many failed attempts.

---

### Refresh Tokens
Rotate refresh tokens and acquire a new access token.

* **URL**: `/api/v1/auth/refresh`
* **Method**: `POST`
* **Content-Type**: `application/json`

#### Request Body
```json
{
  "refresh_token": "AqsclfMk7mvGq2lSTBVs..."
}
```

#### Response (200 OK)
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "refresh_token": "nEb5dih4WEc-AjlFXZi-...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

#### Error Responses
* **401 Unauthorized**: Refresh token is expired or revoked.

---

### Logout User
Revoke the current session and invalidate corresponding tokens.

* **URL**: `/api/v1/auth/logout`
* **Method**: `POST`
* **Headers**:
  * `Authorization: Bearer <access_token>`

#### Response (200 OK)
```json
{
  "message": "logged out"
}
```

#### Error Responses
* **401 Unauthorized**: Access token is invalid, expired, or authorization header format is incorrect.

---

## 2. Diagnostics & System Endpoints

### Health Check
Verify the server process is alive.

* **URL**: `/health`
* **Method**: `GET`

#### Response (200 OK)
```json
{
  "status": "ok"
}
```

---

### Readiness Check
Verify the database connection pool is healthy and ready to accept requests.

* **URL**: `/ready`
* **Method**: `GET`

#### Response (200 OK)
```json
{
  "status": "ready"
}
```

#### Response (503 Service Unavailable)
```json
{
  "error": "Service Unavailable",
  "message": "database not ready"
}
```
