# Contributing to BGScan

Thanks for your interest in contributing. This guide covers the workflow we
follow so reviews stay quick and the codebase stays consistent.

## Getting started

You need Go 1.27 or newer (see `go.mod`).

```bash
git clone --recurse-submodules https://github.com/MohsenBg/bgscan.git
cd bgscan
go build ./...
```

Use `--recurse-submodules`: the repo vendors third-party code under
`third_party/` as submodules.

## Before you change anything

- Check open issues and pull requests first to avoid duplicating work.
- For anything beyond a small fix, open an issue describing the problem and
  your planned approach before writing code.
- Keep pull requests focused on one change. Split unrelated work into
  separate PRs.

## Making changes

- Follow the existing code style. Run `gofmt` on everything you touch.
- Do not edit anything under `third_party/` — that is vendored upstream code.
- Keep comments concise and explain *why*, not *what*. Delete comments that
  just restate the code.
- Update tests alongside behavior changes, and add tests for new behavior.
- Update user-facing docs when behavior changes.

## Verifying your change

Run the same checks CI runs before pushing:

```bash
gofmt -l internal/ cmd/
go vet ./...
go test -race ./...
```

All three must pass. A PR that fails CI will not be merged.

## Tests

```bash
# Full suite with race detector (what CI runs)
go test -race ./...

# A single package while iterating
go test ./internal/core/scanner/...
```

Some probe tests exercise the local network stack (loopback ICMP, UDP/TCP
sockets). They do not need root, but sandboxed environments may restrict
them.

## Commit messages

Write clear, plain-language commit messages in the imperative mood:

- Good: `fix(scanner): cap retry delay at the configured max`
- Bad: `fixed stuff`, `WIP`, `update`

Group related changes into one commit. Do not mix refactoring with behavior
changes in the same commit when it can be avoided.

## Pull requests

- Target the `main` branch.
- Describe what changed and why. Link the issue it closes, if any.
- If the change affects scanning behavior, config format, or the UI,
  explain how you tested it.
- Expect review feedback. Respond to every comment, even if just to agree
  and push a fix.

## Reporting bugs

Open an issue with:

1. What you ran and what you expected.
2. What actually happened (logs help — see `logs/` under the app directory).
3. Your OS, Go version (`go version`), and the BGScan version or commit.

For security vulnerabilities, do not open a public issue — see
[SECURITY.md](SECURITY.md).

## Code of conduct

By participating, you agree to follow the
[Code of Conduct](CODE_OF_CONDUCT.md).
