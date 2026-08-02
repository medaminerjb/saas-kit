# Release Process

This document describes the release process for SaaSKit.

## Versioning

SaaSKit follows [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR** (`X.0.0`): Breaking API changes, major architecture shifts.
- **MINOR** (`0.X.0`): New features, non-breaking additions.
- **PATCH** (`0.0.X`): Bug fixes, security patches, documentation.

## Pre-release Labels

| Label | Meaning | Example |
|-------|---------|---------|
| `-dev` | Development snapshot | `v0.2.0-dev` |
| `-alpha.N` | Feature-complete, not production-tested | `v0.2.0-alpha.1` |
| `-beta.N` | Feature-frozen, hardening in progress | `v0.3.0-beta.2` |
| `-rc.N` | Release candidate, final validation | `v1.0.0-rc.1` |

## Release Checklist

### 1. Prepare

- [ ] All CI checks pass on `main`.
- [ ] `CHANGELOG.md` is updated with all changes since the last release.
- [ ] Version references in documentation are current.
- [ ] Database migrations are tested (up and down).
- [ ] No open security advisories for this release.

### 2. Tag

```bash
# Ensure you are on a clean main branch
git checkout main
git pull origin main

# Tag the release (annotated tag)
git tag -a v0.1.0 -m "v0.1.0: Identity Developer Preview"
git push origin v0.1.0
```

### 3. Build & Publish

- [ ] GitHub Actions CI builds the release artifacts.
- [ ] Docker images are published to the container registry.
- [ ] Go module is available via `go get`.

### 4. Announce

- [ ] Create a GitHub Release with release notes from `CHANGELOG.md`.
- [ ] Post announcement to community channels (Discussions, Slack/Discord).
- [ ] Update the project website if applicable.

## Hotfix Process

For critical security fixes on a released version:

1. Create a branch from the release tag: `git checkout -b hotfix/v0.1.1 v0.1.0`.
2. Apply the fix and add tests.
3. Update `CHANGELOG.md`.
4. Tag and release: `git tag -a v0.1.1 -m "v0.1.1: Security fix for ..."`.
5. Cherry-pick the fix back to `main`.

## Supported Versions

We maintain security patches for the **latest minor release** of each major version. See [SECURITY.md](../SECURITY.md) for the full support matrix.
