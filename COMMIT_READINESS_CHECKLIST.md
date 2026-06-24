# Commit Readiness Checklist

**Before submitting ANY PR to cilium/ebpf, verify this checklist.**

Detailed guide: See [TESTING_BEFORE_PR.md](TESTING_BEFORE_PR.md)

---

## Step 1: Format & Lint Checks ✓

```bash
go tool staticcheck ./...
golangci-lint run ./...
go build -v ./...
```

- [ ] staticcheck: 0 errors
- [ ] golangci-lint: 0 errors
- [ ] build: success

---

## Step 2: Code Generation & Formatting ✓

```bash
make clean && make container-all
git diff --exit-code  # Should have no changes
./scripts/go-fix.sh
git diff --exit-code  # Should have no changes
```

- [ ] Generated files up-to-date
- [ ] No git diff after `make container-all`
- [ ] Go fixes applied
- [ ] No git diff after `go-fix.sh`

---

## Step 2.5: Reset BTF Test Files ⭐

```bash
git checkout -- btf/testdata/
git status  # Should show clean
```

- [ ] BTF test files reset
- [ ] git status shows clean

**Why?** Large BTF test data can cause false diffs if not reset.

---

## Step 3: Unit Tests ✓

```bash
go test -short -count 1 ./...
```

- [ ] All existing tests pass
- [ ] No NEW test failures (pre-existing failures are acceptable)
- [ ] Document any new failures found

---

## Step 4: Race Detector Tests ✓

```bash
go test -race -timeout 5m -short -count 1 ./...
```

- [ ] No race conditions detected
- [ ] Timeout acceptable (5m for short tests)

---

## Step 5: Benchmarks ✓

```bash
go test -short -run '^$' -bench . -benchtime=1x ./...
```

- [ ] Benchmarks run without error
- [ ] No performance regressions (visual check)

---

## Code Quality ✓

- [ ] Comments explain WHY (not WHAT)
- [ ] Exported functions have doc comments
- [ ] No commented-out code (use git history)
- [ ] Import order correct: system → external → internal
- [ ] Error handling standardized (qt.Assert, fmt.Errorf, etc.)

---

## Git Hygiene ✓

- [ ] Commits are focused (single issue/feature per PR)
- [ ] Commit messages are clear and reference issue #
- [ ] Commits are squashed properly (no duplicate changes)
- [ ] Signed-off-by included: `git commit -s`
- [ ] No merge conflicts with main branch
- [ ] Feature branch created from upstream main

---

## Reviewer Engagement ✓

- [ ] Replied to all reviewer comments
- [ ] Addressed all requested changes
- [ ] Asked clarifying questions if needed
- [ ] Provided context for design decisions

---

## Push to GitHub ✓

```bash
git push origin your-branch --force-with-lease
```

- [ ] Branch pushed to remote
- [ ] CI checks running
- [ ] Monitor PR for feedback

---

## Only Commit When ALL Boxes Are Checked ✅

**Do NOT push** if any step above failed locally.

**Quality over speed** — catching issues locally saves PR review cycles.

---

## Full Command Sequence

```bash
#!/bin/bash
set -e

echo "=== Step 1: Format & Lint ==="
go tool staticcheck ./...
golangci-lint run ./...
go build -v ./...

echo "=== Step 2: Code Generation ==="
make clean && make container-all
git diff --exit-code

echo "=== Step 2.5: Go Fixes ==="
./scripts/go-fix.sh
git diff --exit-code

echo "=== Step 2.6: Reset BTF ==="
git checkout -- btf/testdata/
git status

echo "=== Step 3: Unit Tests ==="
go test -short -count 1 ./...

echo "=== Step 4: Race Tests ==="
go test -race -timeout 5m -short -count 1 ./...

echo "=== Step 5: Benchmarks ==="
go test -short -run '^$' -bench . -benchtime=1x ./...

echo "=== ALL CHECKS PASSED ==="
echo "Ready to push!"
```

---

## Troubleshooting

**Golangci-lint cache issues?**
```bash
rm -rf ~/.cache/golangci-lint
golangci-lint run ./...
```

**Staticcheck issues?**
```bash
go clean -cache
go tool staticcheck ./...
```

**Git diff showing unrelated changes?**
```bash
git status
git diff  # Review what changed
git checkout -- <file>  # Reset if needed
```
