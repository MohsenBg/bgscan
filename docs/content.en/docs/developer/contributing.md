---
title: "Contributing"
weight: 6
---
# Contributing Guide

Thank you for your interest in contributing to **bgscan**! Every contribution is appreciated, whether it's fixing bugs, improving documentation, or implementing new features.

## Before You Start

* Search existing Issues before opening a new one.
* If you're planning a large feature or major change, please open an Issue first to discuss it.
* Keep pull requests focused on a single change.

---

## Branch Strategy

Please **do not commit directly to the `main` branch**.

Create a new branch from `main`:

```bash
git checkout main
git pull origin main
git checkout -b feature/my-feature
```

Suggested branch names:

```
feature/add-xray-parser
feature/ui-improvements
fix/memory-leak
fix/windows-build
docs/update-readme
refactor/scanner
test/unit-tests
```

---

## Coding Style

* Follow existing project conventions.
* Keep code simple and readable.
* Prefer small, focused commits.
* Avoid unrelated formatting changes.
* Add comments only when necessary.

---

## Commit Messages

Use descriptive commit messages.

Examples:

```
feat: add HTTP/3 support
fix: resolve race condition in scanner
docs: improve installation guide
refactor: simplify writer pipeline
test: add unit tests for parser
```

---

## Pull Requests

Before opening a Pull Request, ensure:

* Project builds successfully.
* Existing tests pass.
* New functionality is tested.
* Documentation is updated if necessary.

Please include:

* What changed
* Why it changed
* Screenshots (for UI changes)
* Related Issue (if applicable)

---

## Reporting Bugs

Include as much information as possible:

* Operating System
* Architecture
* bgscan version
* Configuration
* Steps to reproduce
* Expected behavior
* Actual behavior
* Logs (if available)

---

## Feature Requests

Please describe:

* The problem you're trying to solve.
* Your proposed solution.
* Possible alternatives.
* Additional context.

---

## Code of Conduct

Be respectful and constructive.

Healthy discussions are encouraged. Personal attacks, harassment, or disrespectful behavior will not be tolerated.

Thank you for helping make bgscan better!

## Related pages

* [Getting Started](../getting-started/) — build and run instructions
* [Architecture](../architecture/) — high-level project layout
* [Core](../core/) — scanner engine, probe interface, config, and result pipeline in detail
* [UI](../ui/) — TUI architecture, component model, layout, and theming
