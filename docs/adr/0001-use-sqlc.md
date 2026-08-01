# ADR-0001: Use sqlc for Database Query Layer

## Status

Accepted

## Context

We need a database query layer for PostgreSQL. Options considered:

1. **GORM** — Full ORM with runtime reflection. Popular but generates unpredictable SQL, has high runtime overhead, and makes complex queries difficult.
2. **sqlx** — Thin wrapper around `database/sql`. Less actively maintained, uses runtime reflection for struct mapping.
3. **Raw pgx** — Direct driver usage. Maximum control but requires manual row scanning, no type safety.
4. **sqlc** — Compile-time code generation from SQL. Type-safe, predictable, zero runtime overhead.

## Decision

Use **sqlc** with **pgx/v5** as the underlying driver.

## Rationale

- **Type safety at compile time** — SQL errors are caught during development, not in production.
- **No runtime reflection** — Generated code is concrete, not reflective.
- **Full SQL control** — We write the exact SQL we want. No ORM query planner surprises.
- **pgx-native** — sqlc generates code targeting pgx directly, leveraging PostgreSQL-specific features (arrays, JSONB, INET, etc.).
- **Zero abstraction cost** — The generated code is as fast as hand-written pgx code.

## Consequences

- Developers must write SQL for all queries (no auto-generated CRUD).
- Schema changes require updating both migration files and sqlc query files.
- `sqlc generate` must be run after changing queries (enforced in CI).
