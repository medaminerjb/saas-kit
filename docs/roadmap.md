# SaaSKit — Architecture & Product Roadmap

> **Vision:** The open-source operating system and infrastructure kernel for SaaS applications — combining the capabilities of Supabase Auth, Clerk, Keycloak, Stripe, and WorkOS into a unified, developer-first Go foundation.

---

## 🏗️ Refined 6-Layer Architecture

The architecture treats developer interfaces (SDKs, CLI, Admin Console) as first-class platform layers rather than secondary UI wrappers, while positioning plugins and Infrastructure as Code as the top-level ecosystem layer.

```
┌─────────────────────────────────────────────────────────────────┐
│  Layer 6 — Ecosystem & Extensibility                           │
│  Compile-Time Plugins · Terraform Provider · App Starter Kits  │
├─────────────────────────────────────────────────────────────────┤
│  Layer 5 — Enterprise Federation & Advanced Auth               │
│  SAML SSO · SCIM 2.0 · LDAP · OpenFGA / ReBAC · Passkeys        │
├─────────────────────────────────────────────────────────────────┤
│  Layer 4 — Developer Platform & Tooling                         │
│  Go SDK · JS/React SDK · CLI (`saaskit`) · Admin Console · Docs │
├─────────────────────────────────────────────────────────────────┤
│  Layer 3 — SaaS Platform Operations                             │
│  Organizations · RBAC · API Keys · Webhooks · Event Engine      │
│  Billing Adapters · Feature Flags & Entitlements               │
├─────────────────────────────────────────────────────────────────┤
│  Layer 2 — Identity & Extensible User Core                      │
│  Users · Auth · Sessions · OIDC Provider · OAuth Federation     │
│  MFA (TOTP) · Extensible Public/Private Metadata (JSONB)       │
├─────────────────────────────────────────────────────────────────┤
│  Layer 1 — Platform Core & Infrastructure                       │
│  Config · Database (`pgx`/`sqlc`) · Migrations · Event Bus      │
│  Envelope Encryption · KMS Adapters · Telemetry Core           │
└─────────────────────────────────────────────────────────────────┘

```

---

## 📊 Master Phase & Version Tracking

| Version Target | Phase | Focus Area | Status | Deliverable Scope |
| --- | --- | --- | --- | --- |
| **`internal`** | **Phase 0** | Foundation | ✅ Completed | Go scaffold, schema, `koanf`, envelope encryption, event bus |
| **`v0.1.0`** | **Phase 1** | Identity | ✅ Completed | Argon2id, OIDC Provider (`zitadel/oidc`), social OAuth2, sessions |
| **`v0.2.0`** | **Phase 2** | Multi-Tenancy | ✅ Completed | Organizations, member invite flows, DB isolation prep, CLI v0.1 |
| **`v0.3.0`** | **Phase 3** | Authorization & MFA | 🟢 **OPEN (Next)** | Tenant-scoped RBAC, TOTP, recovery codes, token grace rotation |
| **`v0.4.0`** | **Phase 4** | Metadata & Extensible Identity | 🟡 **OPEN** |  JSONB metadata, schema validation, GIN indexes, metadata events |
| **`v1.0.0`** | **Phase 5** | SaaSKit Core GA | 🟡 **OPEN** | Tenant API Keys, Public Event Schema, Admin Console, SDKs, KMS |
| **`v1.2.0`** | **Phase 6** | Integration Platform | 🟡 **OPEN** | Outbound Webhook Engine, OAuth App Provider, Email Abstraction |
| **`v1.5.0`** | **Phase 7** | Enterprise Federation | 🟡 **OPEN** | SAML 2.0, Home Realm Discovery, SCIM 2.0, LDAP Directory Sync |
| **`v1.7.0`** | **Phase 8** | SaaS Operations & Billing | 🟡 **OPEN** | Stripe adapter, feature flags engine, usage metering aggregators |
| **`v2.0.0`** | **Phase 9** | Advanced Isolation & ReBAC | 🟡 **OPEN** | PostgreSQL RLS, OpenFGA Connector, Passkeys (WebAuthn) |
| **`v2.2.0`** | **Phase 10** | Infrastructure Maturity | 🟡 **OPEN** | Active-Active HA, multi-region routing, PgBouncer optimization |
| **`v2.5.0`** | **Phase 11** | Compliance & Residency | 🟡 **OPEN** | GDPR data export/erasure, SOC2 evidence, data residency pinning |
| **`v3.0.0`** | **Phase 12** | Ecosystem & Starter Kits | 🟡 **OPEN** | Go Plugin SDK, Terraform Provider, `saaskit create` starter templates |

---

## 🗺️ Master Deliverables & Checklist

### Phase 0 — Foundation ✅

* **Status:** Completed · **Release:** `internal`

* [x] **Go project scaffold** — `chi`, `pgx`, `sqlc`, `koanf`, `slog`
* [x] **PostgreSQL schema** — 10 base migrations
* [x] **Configuration system** — Environment variables + YAML validation
* [x] **Connection pool** — Configurable `pgx` pool with health check
* [x] **Envelope encryption** — AES-256-GCM with HKDF key derivation
* [x] **Event publisher interface** — Log, Multi, and Noop implementations
* [x] **Background job scheduler** — Graceful ticker shutdown system
* [x] **Containerization** — Docker Compose + multi-stage Dockerfiles
* [x] **CI pipeline** — GitHub Actions lint, test, and build suite
* [x] **ADRs** — Architecture decisions (`sqlc`, asymmetric JWTs, `zitadel/oidc`)
* [x] **Open-source license** — Apache 2.0

---

### Phase 1 — Identity ✅

* **Status:** Completed · **Release:** `v0.1.0`

* [x] **User domain model** — 6-state lifecycle (`active`, `disabled`, `locked`, `pending`, `invited`, `deleted`)
* [x] **Argon2id password hashing** — PHC format with configurable memory/threads
* [x] **JWT key management** — RS256/ES256/EdDSA dynamic key handling
* [x] **Auth service core** — Registration, login, refresh rotation, logout
* [x] **Password reset & email verification** — Generic identity token flows
* [x] **Session management** — List, revoke by ID, and automated background cleanup
* [x] **OIDC Provider** — Full `zitadel/oidc` implementation with PKCE support
* [x] **Social login** — Google & GitHub OAuth2 relying party integration
* [x] **Login/Consent UI** — Dark-themed Go HTML templates
* [x] **Audit event logging** — System audit trail for all identity actions

---

### Phase 2 — Multi-Tenancy ✅

* **Status:** Completed · **Release:** `v0.2.0`

* [x] **Tenant domain model** — Organization attributes (`id`, `name`, `slug`, `status`)
* [x] **Organization CRUD** — Multi-org membership per user
* [x] **User invitations** — Email token workflow (`accepted`, `rejected`)
* [x] **Tenant context middleware** — Dynamic extraction of `tenant_id` from JWT to set DB session vars
* [x] **Tenant-scoped queries** — Strict repository-level query isolation
* [x] **Tenant connection resolver** — Connection router interface for future DB-per-tenant support
* [x] **Developer CLI (v0.1)** — Local user setup, tenant management, and migration execution

---

### Phase 3 — Authorization & MFA ✅

* **Status:** Completed · **Release:** `v0.3.0`

* [x] **Role model** — Pre-configured system roles (`Owner`, `Admin`, `Manager`, `Member`, `Viewer`)
* [x] **Granular permission model** — Resource-action formatting (`tenant.read`, `tenant.update`, `members.invite`, `members.remove`)
* [x] **Permission middleware** — Route protection via `RequirePermission("resource.action")`
* [x] **Tenant-scoped RBAC** — Contextual roles bound per organization
* [x] **MFA Framework** — Enrollment, challenge verification, and recovery systems
* [x] **TOTP Provider** — Authenticator app setup, encrypted secret storage, backup codes
* [x] **Graceful token rotation** — 10-second grace window to eliminate concurrent refresh token race conditions

---

### Phase 4 — Metadata & Extensible Identity 🟢

* **Status:** OPEN (Immediate Priority) · **Release:** `v0.4.0`

* [ ] **User Metadata Storage** — `metadata_public` (client-accessible) and `metadata_private` (backend-only) JSONB columns on `users`
* [ ] **Organization Metadata Storage** — `metadata` JSONB column on `tenants` for billing IDs, locales, and custom configs
* [ ] **GIN Indexing & Validations** — JSON schema validation and size cap enforcement (32KB max per payload)
* [ ] **Metadata RBAC Rules** — Permission checks securing read/write metadata actions across tenant boundaries
* [ ] **Metadata CRUD API Endpoints** — Dedicated REST endpoints (`/api/v1/users/{id}/metadata`, `/api/v1/tenants/{id}/metadata`)
* [ ] **Metadata Event Stream** — Publish `user.metadata.updated` and `organization.metadata.updated` system events

---

### Phase 5 — SaaSKit Core GA 🟡

* **Status:** OPEN · **Release:** `v1.0.0` 🎉

* [ ] **Tenant-Aware API Keys** — Key prefixing (`sk_live_...`), scopes, and automatic context extraction:

$$\text{API Key} \xrightarrow{\quad\text{Validation}\quad} \text{Tenant Context} \xrightarrow{\quad\text{RBAC}\quad} \text{Permission Enforced}$$


* [ ] **Public System Event Model** — Structured JSON event format (`event`, `tenant_id`, `actor`, `timestamp`, `data`)
* [ ] **Go SDK (`saaskit-go`)** — Strongly typed client library with retry controls and backoff algorithms
* [ ] **JavaScript SDK (`saaskit-js`)** — Client library with login components and metadata helpers
* [ ] **Admin Console (`v1.0`)** — Web UI (React + Vite) for managing users, tenants, keys, metadata, and audit logs
* [ ] **Machine-to-Machine (M2M) Auth** — OAuth2 Client Credentials grant flow
* [ ] **Cloud KMS Adapters** — Key management integration for AWS KMS, GCP KMS, and HashiCorp Vault
* [ ] **OpenTelemetry Integration** — Prometheus metrics and OTLP distributed tracing collectors
* [ ] **JWKS Key Rotation Engine** — Signature key rotation without invalidating active sessions
* [ ] **Starter Template Generator (v0.1)** — Early version of `saaskit create` for instant local project bootstrapping
* [ ] **Documentation Platform (`docs.saaskit.dev`)** — Quickstarts, deployment guides, and `examples/` repo (`basic-saas`, `multi-tenant-saas`)

---

### Phase 6 — Integration Platform 🟡

* **Status:** OPEN · **Release:** `v1.2.0`

* [ ] **Webhook Engine** — Dynamic subscription endpoints, asynchronous worker pool, and delivery queues
* [ ] **Cryptographic Signatures** — HMAC-SHA256 headers for webhook verification and replay prevention
* [ ] **OAuth Application Provider** — Allow third-party applications to build integrations against SaaSKit
* [ ] **Email Provider Abstraction** — Mailer drivers for SendGrid, Resend, Postmark, and SMTP

---

### Phase 7 — Enterprise Federation 🟡

* **Status:** OPEN · **Release:** `v1.5.0`

* [ ] **SAML 2.0 SP Implementation** — Enterprise IdP integration (Okta, Azure AD, Ping Identity)
* [ ] **Home Realm Discovery (HRD)** — Automatic domain matching to route users to designated enterprise IdPs
* [ ] **SCIM 2.0 Provisioning** — Inbound user and group provisioning/deprovisioning APIs
* [ ] **LDAP Synchronization** — Active Directory and LDAP server user synchronization
* [ ] **Attribute Mapper** — Mapping SAML, SCIM, and LDAP attributes into dynamic user metadata

---

### Phase 8 — SaaS Operations & Billing 🟡

* **Status:** OPEN · **Release:** `v1.7.0`

* [ ] **Stripe Adapter** — Subscription state tracking, plan synchronization, and webhook verification
* [ ] **Billing Metadata Linkage** — Store customer IDs, subscription statuses, and invoice links directly in tenant metadata
* [ ] **Feature Flags Engine** — Rule-based flag evaluation using user and tenant metadata attributes
* [ ] **Usage Metering Engine** — Time-series counter aggregations for seats, storage, and API usage

---

### Phase 9 — Advanced Isolation & ReBAC 🟡

* **Status:** OPEN · **Release:** `v2.0.0`

* [ ] **PostgreSQL Row-Level Security (RLS)** — Shared-database RLS policies and execution templates
* [ ] **Schema-per-Tenant Isolation** — Dynamic schema routing via the tenant connection resolver
* [ ] **Database-per-Tenant Isolation** — Multi-database dynamic connection routing and migration runner
* [ ] **OpenFGA Connector** — Plug-and-play adapter for relationship-based authorization (ReBAC)
* [ ] **Passkeys (WebAuthn)** — Biometric FIDO2 passwordless authentication

---

### Phase 10 — Infrastructure Maturity 🟡

* **Status:** OPEN · **Release:** `v2.2.0`

* [ ] **High Availability Architecture** — Active-Active cluster deployments with `PgAdvisory` lock management
* [ ] **Multi-Region Routing** — Geo-distributed data routing and read replica optimization
* [ ] **PgBouncer Integration** — Prepared statement connection pooling configurations

---

### Phase 11 — Compliance & Residency 🟡

* **Status:** OPEN · **Release:** `v2.5.0`

* [ ] **GDPR Automation** — Data export tools and automated right-to-be-forgotten deletion workflows
* [ ] **SOC2 Evidence Engine** — Automated report generation for audit logging and access control rules
* [ ] **Data Residency Pinning** — Geographically bound database record placement rules

---

### Phase 12 — Ecosystem & Starter Kits 🟡

* **Status:** OPEN · **Release:** `v3.0.0`

* [ ] **Compile-Time Plugin SDK** — Interface-based Go plugins for custom billing, storage, and notification drivers
* [ ] **Terraform Provider** — Infrastructure as Code provider for managing tenants, OIDC clients, keys, and roles
* [ ] **CLI Scaffolding Framework (`saaskit create`)** — Full starter generators for production SaaS applications

---

## 📐 Strategic Architecture & Open Decisions

| ID | Topic | Target Phase | Recommendation | Status |
| --- | --- | --- | --- | --- |
| **#1** | **Billing Scope** | Phase 8 (`v1.7.0`) | Thin event relay only (Stripe first, defer Paddle/LemonSqueezy) | 🟡 **OPEN** |
| **#2** | **Fine-Grained Auth** | Phase 9 (`v2.0.0`) | Do not build internal Zanzibar engine; build official OpenFGA & Casbin connectors | 🟡 **OPEN** |
| **#3** | **Tenant Connection Routing** | Phase 2/9 (`v2.0.0`) | Expose repository connection resolver interface from Phase 2; plug in dynamic DB routing in Phase 9 | 🟡 **OPEN** |
| **#4** | **Plugin System** | Phase 12 (`v3.0.0`) | Build compile-time Go plugins (Caddy style) rather than WASM sandboxing | 🟡 **OPEN** |

---

## 🎯 Production Performance & Target SLA

| Metric Category | Target Value | Verification Standard |
| --- | --- | --- |
| **Tenant Scale** | $1,000$ tenants (v1.0) $\rightarrow 100,000$ tenants (v3.0) | Multi-tenant schema indexing tests |
| **User Capacity** | $100,000$ users (v1.0) $\rightarrow 10,000,000$ users (v3.0) | Load testing against PostgreSQL |
| **Authentication Latency** | $p_{95} < 100\text{ ms}$ (v1.0) $\rightarrow p_{95} < 25\text{ ms}$ (v3.0) | End-to-end HTTP bench |
| **Token Verification Speed** | $< 5\text{ ms}$ local verification | In-memory asymmetric key checks |
| **Throughput Target** | $10,000\text{ RPS}$ concurrent auth requests | CPU/Memory optimization benchmarking |
| **Audit Event Volume** | $10,000,000$ events (v1.0) $\rightarrow 1,000,000,000$ events (v3.0) | DB partition scaling tests |