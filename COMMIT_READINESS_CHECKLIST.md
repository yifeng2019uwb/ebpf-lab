# Commit Readiness Checklist

**Before submitting ANY PR to cilium/ebpf, verify this checklist.**

## Code Quality

- [ ] Code compiles: `go build -v ./...`
- [ ] No staticcheck errors: `go tool staticcheck ./...`
- [ ] No golangci-lint errors: `golangci-lint run ./...`
- [ ] Imports formatted: `goimports` (handled by golangci-lint)

## Testing

- [ ] All existing unit tests pass: `go test -race -timeout 5m -short -count 1 ./...`
- [ ] No NEW test failures introduced (pre-existing failures are acceptable)
- [ ] Issue-specific tests pass (if applicable)

## Code Generation

- [ ] Generated files up-to-date: `make clean && make container-all`
- [ ] No git diff after generation
- [ ] Go fixes applied: `./scripts/go-fix.sh`
- [ ] No git diff after fixes

## Documentation

- [ ] Comments explain WHY (not WHAT)
- [ ] Exported functions have doc comments
- [ ] No commented-out code (use git history)

## Git

- [ ] Changes are focused (single issue/feature per PR)
- [ ] Commit messages are clear
- [ ] No merge conflicts with main branch

## Only Commit When ALL Boxes Are Checked ✅
