---
title: "Contributing"
weight: 6
---

# Contributing guide

Thank you for wanting to help **BGScan**. Whether you fix a bug, improve the docs, or add a new feature, your contribution helps the project get better.

## Before you start

- Check the existing issues before opening a new one, to make sure the topic has not come up already.
- If you plan to add a large feature or make an important change, it is better to open an issue first and talk about the proposed solution.
- Try to keep each Pull Request focused on one specific change.

---

## Branch strategy

Please **do not commit directly on `main`**. Create a separate branch for your changes:

```bash
git checkout main
git pull origin main
git checkout -b feature/my-feature
```

You can use these patterns for branch names:

```text
feature/add-xray-parser
feature/ui-improvements
fix/memory-leak
fix/windows-build
docs/update-readme
refactor/scanner
test/unit-tests
```

---

## Coding style

- Follow the conventions and patterns already in the project.
- Write simple, readable code.
- Break changes into small, specific commits.
- Avoid formatting changes unrelated to your work.
- Add comments only where they help the code read better.

---

## Commit messages

A commit message should be short and specific and explain what changed.

A few examples:

```text
feat: add HTTP/3 support
fix: resolve race condition in scanner
docs: improve installation guide
refactor: simplify writer pipeline
test: add unit tests for parser
```

---

## Pull Request

Before opening a Pull Request, check these items:

- The project builds without errors.
- The existing tests pass.
- A new feature has suitable tests when needed.
- The related documentation has been updated.

In the Pull Request description, please also write:

- What changed?
- Why is this change needed?
- If the interface changed, add a suitable screenshot or image.
- If there is a related issue, link to it.

---

## Reporting a bug

To make the bug easier to review and fix, include as much of this information as you can in the issue:

- Operating system
- System architecture
- BGScan version
- Relevant settings
- Steps to reproduce the bug
- Expected behavior
- Actual program behavior
- Relevant logs, if available

The more precise the information you give, the easier it is to find the cause of the problem.

---

## Feature request

If you want to suggest a new feature, it is better to explain these items:

- What problem do you have and what will this feature solve?
- What solution do you suggest?
- Is there another alternative solution?
- What other information or details could help review the suggestion?

---

## Code of conduct

Treat others with respect and be constructive. We welcome healthy discussion, but personal insults, harassment, and disrespectful behavior are not acceptable.

Thank you for taking the time to make BGScan better!

## Related pages

- [Getting started](./getting-started/): build and run commands for the project
- [Architecture](./architecture/): an introduction to the overall project structure
- [Core](./core/): details of the scan engine, the Probe interface, settings, and result processing
- [User interface](./ui/): TUI architecture, components, layout, and themes
