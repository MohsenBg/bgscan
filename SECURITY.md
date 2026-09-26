# Security Policy

## Reporting a vulnerability

Do not open a public issue for security vulnerabilities. Report them
privately to moh.1380.1393@gmail.com with:

1. A description of the vulnerability and its potential impact.
2. Steps to reproduce, if possible.
3. The BGScan version or commit you tested against.

You will get an initial response within a week. If the report is confirmed,
a fix will be prepared and released before any public disclosure, and you
will be credited unless you prefer otherwise.

## Supported versions

Only the latest release on the `main` branch receives security fixes.
Older releases are not patched — please upgrade before reporting.

## Scope

In scope: the code in this repository under `cmd/` and `internal/`
(the scanner, its probes, and the UI).

Out of scope: vendored code under `third_party/` (report those issues
upstream), and the documentation site under `docs/`.
