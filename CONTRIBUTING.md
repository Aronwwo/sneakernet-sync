# Contributing to sneakernet-sync

Thank you for contributing! Please read these guidelines before submitting a PR.

---

## Branch Naming

| Prefix        | Purpose                              |
|---------------|--------------------------------------|
| `feature/`    | New feature or enhancement           |
| `fix/`        | Bug fix                              |
| `test/`       | Tests only                           |
| `docs/`       | Documentation changes                |
| `refactor/`   | Refactoring without behavior change  |
| `ci/`         | CI/CD changes                        |

Example: `feature/conflict-detection`, `fix/scan-hidden-files`

---

## Conventional Commits

All commit messages must follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>: <short description>
```

Types: `feat`, `fix`, `test`, `docs`, `refactor`, `ci`, `chore`

Examples:
```
feat: add three-way reconciliation
fix: skip hidden directories in scanner
test: add coverage for hash package
docs: update README quick start
```

---

## Pull Request Process

1. Create a branch from `main` using the naming convention above.
2. Implement your changes with tests.
3. Ensure all CI checks pass locally:
   ```bash
   make all
   ```
4. Open a PR against `main`.
5. Request a review from at least one team member.
6. Address review feedback.
7. CI must pass (all platforms, lint, build) before merging.
8. Squash-merge into `main`.

---

## Definition of Done

See [docs/DEFINITION_OF_DONE.md](docs/DEFINITION_OF_DONE.md) for the full checklist that every change must satisfy before merging.

---

## Code Style

- Run `gofmt` and `goimports` before committing.
- All exported symbols must have doc comments.
- Cyclomatic complexity must stay below 15 per function.
- No `//nolint` without a comment explaining why.

---

## Review Requirements

- At least one approval from a team member required.
- All CI jobs must be green.
- No unresolved review comments.
