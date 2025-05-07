# Contributing Guide

Welcome! We appreciate your interest in contributing to this project.  
Please review the guidelines below before making a contribution.

## 📝 Enhancement Proposals (EPs)

For significant changes or new features:

1. Open a new **Issue**.
2. Choose the **Enhancement Proposal** template.
3. Fill out the required sections (Summary, Motivation, Proposal Details, etc.).
4. Proposals will be discussed and reviewed before acceptance.

Accepted proposals are tracked in [`enhancement-proposals/`](./enhancement-proposals/).

## 🔀 Pull Requests

When submitting a Pull Request (PR):

1. Choose the correct template:
    - `Feature`
    - `Bugfix`
    - `Chore`
    - `Documentation`
2. Follow the checklist included in the template.
3. Reference any related Issue or Enhancement Proposal.
4. Ensure your branch is named appropriately:
    - `feature/your-feature-name`
    - `bugfix/your-bug-description`
    - `chore/task-description`

## ✅ Code Style

- Follow Go conventions (`go fmt` before committing).
- Write clear commit messages (`<type>: <short description>`).
- Refer to the [Effective Go](https://go.dev/doc/effective_go) style guide for idiomatic Go code.

## 🔄 Continuous Integration

Our CI workflow runs automatically on all branches. Ensure your changes:
- Pass `go build ./...`
- Pass `go test ./...`

## ❓ Need Help?

If you have questions about the process, templates, or anything else, feel free to open a discussion or reach out to a maintainer.

---

Thanks for contributing and helping improve the project!
