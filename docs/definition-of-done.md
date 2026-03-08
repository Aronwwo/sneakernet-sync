# Definition of Done

Every completed feature, iteration, or deliverable must satisfy these criteria:

## Code Quality

- [ ] Code compiles without errors (`go build ./...`)
- [ ] No lint warnings (`golangci-lint run`)
- [ ] No race conditions (`go test -race ./...`)
- [ ] No obvious placeholders or unfinished TODO in critical paths

## Testing

- [ ] Unit tests exist for new/changed functionality
- [ ] All tests pass
- [ ] Key edge cases are covered
- [ ] Integration tests cover the primary workflow

## Documentation

- [ ] Code has Go doc comments on exported types and functions
- [ ] Changes are reflected in relevant docs
- [ ] Architecture decisions are recorded in ADRs
- [ ] Assumptions are documented

## Review

- [ ] Code is readable and follows Go conventions
- [ ] Variable and function names are clear
- [ ] Error handling is explicit
- [ ] No silent data loss paths

## CI/CD

- [ ] CI pipeline passes (build + test + lint)
- [ ] Cross-platform compatibility verified (Linux, Windows, macOS)
