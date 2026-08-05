# Contributing to SaaSKit

Thank you for your interest in contributing to SaaSKit! As a CNCF-aligned, open-source project, we welcome contributions of all forms: bug fixes, features, documentation, design suggestions, or community help.

Before contributing, please read this document and our [Code of Conduct](CODE_OF_CONDUCT.md).

---

## 1. Developer Certificate of Origin (DCO)

We require all contributions to be certified under the Developer Certificate of Origin. This means you must sign off your git commits to confirm you have the right to submit the code under the project's Apache 2.0 license.

To sign off on a commit:
```bash
git commit -s -m "pkg/component: descriptive commit message"
```
If you forget to sign off, you can amend your last commit:
```bash
git commit --amend --signoff
```
Or do an interactive rebase to sign off multiple historical commits.

---

## 2. Commit Message Guidelines

We follow the **Conventional Commits** specification for commit messages. This helps automate releases and changelogs.

Format:
```text
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

* **Types:**
  - `feat`: A new feature
  - `fix`: A bug fix
  - `docs`: Documentation changes
  - `style`: Code style changes (formatting, missing semi-colons, no functional changes)
  - `refactor`: Code changes that neither fix bugs nor add features
  - `test`: Adding or correcting tests
  - `chore`: Infrastructure, dependencies, or build tasks
* **Scope:** The package or module name (e.g., `auth`, `oidc`, `database`, `config`).
* **Example:**
  ```text
  feat(auth): implement refresh token rotation with HMAC hashing
  
  Encrypt raw refresh tokens on write and verify them using the server secret.
  
  Close #42
  Signed-off-by: Jane Doe <jane.doe@example.com>
  ```

---

## 3. Development Setup

### Prerequisites
- **Go:** Version 1.22 or higher.
- **Docker & Docker Compose:** For running PostgreSQL.
- **Make:** For build and migration tasks.

### Getting Started
1. **Fork and Clone the repository:**
   ```bash
   git clone https://github.com/YOUR-USERNAME/saaskit.git
   cd saaskit
   ```
2. **Setup Local Configuration:**
   ```bash
   cp .env.example .env
   # Customize .env if needed (default values work out of the box with Docker Compose)
   ```
3. **Start PostgreSQL:**
   ```bash
   docker compose up -d db
   ```
4. **Run Database Migrations:**
   ```bash
   make migrate-up
   ```
5. **Generate JWT signing keys (for development):**
   ```bash
   make keys
   ```
6. **Run SaaSKit locally:**
   ```bash
   make run-direct
   ```

---

## 4. Coding Standards

- **Formatting:** All Go code must be formatted using `gofmt`. Run `make lint` before submitting a PR.
- **Linting:** We use `golangci-lint` to enforce style and detect common errors.
- **Testing:** Add unit tests for all new services, repositories, and helper functions. Ensure tests run with the race detector enabled (`go test -race ./...`).

## 5. Security Expectations

- **Vulnerability Scanning:** All code must pass `govulncheck` without critical or high-severity vulnerabilities.
- **Secrets Management:** Never commit secrets, API keys, or credentials to the repository. Use environment variables or secret managers.
- **Dependencies:** Keep dependencies up-to-date. Security-sensitive dependencies should be pinned by commit SHA in CI workflows.
- **Code Review:** Security-related changes require additional review. Be mindful of authentication, authorization, and data handling.
- **Fuzz Testing:** For security-critical components (crypto, parsing, validation), consider adding fuzz tests.

---

## 6. Pull Request Process

1. **Create a Branch:** Keep your branches focused. Create a branch from `master` with a meaningful name:
   ```bash
   git checkout -b feat/oauth-providers
   ```
2. **Implement Your Changes:** Write code, add tests, and update documentation as needed.
3. **Run Checks Locally:**
   ```bash
   make lint
   make test
   ```
4. **Submit PR:** Push your branch to your fork and submit a PR to the `master` branch.
5. **Review & Revisions:** A maintainer will review your PR. Address any review comments by pushing commits to your branch.
6. **Merge:** Once approved and all CI checks pass, a maintainer will merge your PR.
