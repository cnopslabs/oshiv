# Enhancement Proposal 0001: Establishing Proposal Process, Templates, and Repository Housekeeping

## Summary

This proposal introduces a structured process for submitting, reviewing, and tracking Enhancement Proposals (EPs) in this project. It also establishes contribution templates, sets up continuous integration (CI), and implements repository housekeeping practices, including the removal of the `vendor` directory from Git history.

## Motivation

As the project grows, a consistent and transparent process for proposing and discussing significant changes is essential. Enhancement Proposals will help the team:

- Document ideas and changes formally.
- Provide context and rationale for decisions.
- Enable collaborative review and discussion.

Additionally, removing the tracked `vendor` directory reduces repository size and promotes best practices by using Go modules for dependency management.

## Proposal Details

This proposal includes the following changes:

- Created the `enhancement-proposals/` directory to store all EPs.
- Added an **Enhancement Proposal issue template** at `.github/ISSUE_TEMPLATE/enhancement-proposal.md`.
- Added multiple **Pull Request templates** at:
    - `.github/PULL_REQUEST_TEMPLATE/feature.md`
    - `.github/PULL_REQUEST_TEMPLATE/bugfix.md`
    - `.github/PULL_REQUEST_TEMPLATE/chore.md`
    - `.github/PULL_REQUEST_TEMPLATE/documentation.md`
- Implemented a **GitHub Actions workflow** (`.github/workflows/build.yml`) tailored for Go projects.
- Configured the workflow to trigger on all branches (`*`), while recommending clear branch naming:
    - `main`
    - `develop`
    - `feature-*` or `feature/*`
- Updated `.gitignore` and removed the `vendor` directory from Git history.
- Updated `README.md` to reference the Enhancement Proposal process.

## Alternatives Considered

- **No formal proposal process** — would lead to ad-hoc changes, poor documentation, and potential miscommunication.
- **Keeping the vendor directory** — rejected to avoid bloating the repository and to maintain best practices for Go dependency management.

## Risks / Drawbacks

Minimal. The introduction of templates and workflow automation adds slight initial overhead but greatly improves long-term collaboration and maintainability.

## Adoption Plan

- Merge this proposal and associated repository changes.
- Communicate the new processes and templates to all contributors.
- Use the Enhancement Proposal process for all significant future changes.

## References

- [Go Modules Documentation](https://golang.org/ref/mod)