# Definition of Done

Every task, user story, or pull request must satisfy **all** of the following criteria before it can be considered done and merged into `main`.

---

## Checklist

- [ ] **Code works** — the feature or fix behaves as described in the acceptance criteria
- [ ] **Compiles** — `go build ./...` succeeds with no errors
- [ ] **Tests pass** — `go test ./...` passes with no failures
- [ ] **Coverage ≥ 80%** — test coverage for new/changed code is at least 80%
- [ ] **Code reviewed** — at least one team member has approved the PR
- [ ] **CI green** — all GitHub Actions jobs pass (test on Linux/Windows/macOS, lint, build)
- [ ] **Lint clean** — `golangci-lint run` produces no new warnings or errors
- [ ] **No regressions** — existing tests still pass; no previously working functionality is broken
- [ ] **Merged to main** — the branch has been merged (squash merge) into `main`
- [ ] **Documentation updated** — any public API, CLI command, or user-facing behaviour change is reflected in the README or relevant docs

---

## Notes

- Coverage target applies to non-placeholder code. Placeholder functions (`return fmt.Errorf("not implemented yet")`) are excluded from coverage requirements until implemented.
- If a CI job is flaky due to infrastructure (not code), the team lead may approve a merge after manual verification, but the flakiness must be tracked as a follow-up issue.
