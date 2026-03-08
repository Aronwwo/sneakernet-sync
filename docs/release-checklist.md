# Release Checklist

## Pre-Release

- [ ] All tests pass (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] Cross-platform build succeeds (`make build` for Linux, Windows, macOS)
- [ ] Integration tests pass (two-device sync scenario)
- [ ] Documentation is up to date
- [ ] CHANGELOG updated
- [ ] Version number updated

## Manual Verification

- [ ] `init` creates metadata directory correctly
- [ ] `scan` detects all file types (files, directories)
- [ ] `status` shows correct information
- [ ] `push` creates valid USB media structure
- [ ] `pull` imports and applies changes correctly
- [ ] `sync` performs full cycle
- [ ] `conflicts` lists unresolved conflicts
- [ ] `resolve` marks conflicts as resolved
- [ ] `doctor` reports integrity status
- [ ] `--dry-run` prevents any file writes
- [ ] Hidden files are properly skipped
- [ ] Conflict detection works for all 7 rules

## Release

- [ ] Tag version in git
- [ ] Build binaries for all platforms
- [ ] Create GitHub release with binaries
- [ ] Update README with latest version

## Post-Release

- [ ] Verify published binaries work
- [ ] Update project board / backlog
