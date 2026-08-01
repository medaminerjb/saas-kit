# ADR-0003: Use koanf for Configuration Management

## Status

Accepted

## Context

The application needs structured configuration from multiple sources (environment variables, YAML files, defaults). Options considered:

1. **Viper** — Feature-rich but heavy, uses global state, complex dependency tree.
2. **koanf** — Lightweight, composable, explicit. No global state.
3. **envconfig** — Environment-only. Too limited for YAML fallback.

## Decision

Use **koanf** with environment variable provider (prefix `SAASKIT_`) and optional YAML file fallback.

## Rationale

- **No global state** — Each config instance is explicit, making testing straightforward.
- **Small dependency footprint** — Minimal transitive dependencies compared to Viper.
- **Composable providers** — env, YAML, defaults can be layered cleanly with clear precedence.
- **12-factor compatible** — Environment variables take precedence over file-based config.

## Consequences

- Configuration struct must be defined with explicit field mappings.
- Less "magic" than Viper — requires slightly more boilerplate but is more predictable.
