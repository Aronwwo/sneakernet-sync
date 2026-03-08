# Git Workflow

This document describes the branching strategy and collaboration workflow used in sneakernet-sync.

---

## Branches

| Branch       | Purpose                                          |
|--------------|--------------------------------------------------|
| `main`       | Stable, always-deployable code. Protected.       |
| `feature/*`  | New features and enhancements                    |
| `fix/*`      | Bug fixes                                        |
| `test/*`     | Test-only changes                                |
| `docs/*`     | Documentation changes                            |
| `refactor/*` | Refactoring without observable behaviour change  |
| `ci/*`       | CI/CD pipeline changes                           |

**Never commit directly to `main`.** All changes go through a Pull Request.

---

## Feature Branch Workflow

1. **Branch off `main`:**
   ```bash
   git checkout main
   git pull origin main
   git checkout -b feature/your-feature-name
   ```

2. **Commit small, focused changes** using [Conventional Commits](https://www.conventionalcommits.org/):
   ```
   feat: add conflict detection for concurrent edits
   fix: correct off-by-one in WalkDir skip logic
   test: add edge cases for empty directory scan
   ```

3. **Push and open a PR:**
   ```bash
   git push origin feature/your-feature-name
   # Then open a PR on GitHub targeting main
   ```

4. **CI must pass** — all matrix jobs (Linux/Windows/macOS, Go 1.22/1.23), lint, and build must be green.

5. **One approval required** from the other team member.

6. **Squash merge** into `main` — keep `main` history linear.

7. **Delete the branch** after merging.

---

## Sprint Tags

At the end of each sprint, tag `main` with the sprint number:

```bash
git tag sprint-N
git push origin sprint-N
```

Example: `sprint-0`, `sprint-1`, `sprint-2`, …

---

## Conventional Commits Reference

| Type        | When to use                                          |
|-------------|------------------------------------------------------|
| `feat`      | A new feature                                        |
| `fix`       | A bug fix                                            |
| `test`      | Adding or updating tests                             |
| `docs`      | Documentation only                                   |
| `refactor`  | Code change that neither fixes a bug nor adds a feature |
| `ci`        | CI configuration changes                             |
| `chore`     | Maintenance tasks (dependency bumps, cleanup)        |

---

## Conflict Resolution

If your branch has conflicts with `main`:

```bash
git fetch origin
git rebase origin/main
# Resolve conflicts, then:
git rebase --continue
git push --force-with-lease origin feature/your-feature-name
```
