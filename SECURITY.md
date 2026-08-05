# Security Policy

We take the security of SaaSKit seriously. If you believe you have found a security vulnerability, please report it to us responsibly.

## Supported Versions

We support security updates for the following versions:

| Version | Supported | Notes |
|---------|-----------|-------|
| v0.5.x  | ✅ Yes    | Current Release |
| v0.4.x  | ✅ Yes    | Previous Release |
| < v0.4.0| ❌ No     | End of Life |

We actively backport security fixes to the latest patch release of each supported minor version.

## Reporting a Vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Instead, please report security issues privately:
1. Email your report to **security@saaskit.dev**.
2. If possible, encrypt your email using our PGP key (available on key servers, Fingerprint: `D3B0 73E5 111E 61C1 A534 82D4 59AA BC4D 2E3F 991F`).
3. Include as much detail as possible in your report, including:
   - A descriptive title.
   - The affected components, endpoints, or functions.
   - Steps to reproduce the issue (including proof-of-concept scripts or API payloads where applicable).
   - The potential impact of the vulnerability.

We will acknowledge receipt of your report within 24 hours and provide an initial assessment within 3 business days.

## Disclosure Process

SaaSKit follows a coordinated disclosure process:
1. **Investigation:** We will investigate the issue and verify the vulnerability.
2. **Fix Development:** We will develop a fix/patch privately.
3. **Notification:** We will request a CVE identifier and prepare release notes.
4. **Release:** We will release the fix in a new patch release.
5. **Public Disclosure:** We will publish security advisory details 90 days after the initial report, or immediately upon releasing the patch if coordinated with the reporter.

## Response SLA

* **Acknowledgment:** Within 24 hours.
* **Triage & Classification:** Within 3 business days.
* **Embargo Duration:** Up to 90 days (standard industry practice).
* **Credit:** We will publicly credit you in our security advisories and release notes (unless you prefer to remain anonymous).
