# Contributing to cilium/ebpf

## Process

1. **Find an issue or gap** — read open issues, use the library, notice something wrong/missing
2. **Discuss first** — Cilium Slack `#ebpf-go-dev` (ebpf.io/slack) or open a GitHub Discussion
   - Key rule: less new API = easier to merge
   - Don't code blind — align with maintainers before writing
3. **Write code + tests** — PR must pass CI and have tests
4. **Sign off every commit** — DCO required (see below)
5. **Open PR** — must be ready to merge (not draft unless asking for early feedback)

## DCO — Developer Certificate of Origin

Every commit must have `Signed-off-by` footer. Means: "I wrote this and have the right to contribute it."

```bash
# automatic — always use -s flag:
git commit -s -m "your message"

# result:
# fix: improve error message in map loader
#
# Signed-off-by: Yifeng Zhang <yifeng2019@gmail.com>
```

Without it → CI bot rejects PR automatically.

## Commit Message Format

Follow the pattern used in the repo:

```
<area>: <what changed>          ← general change
fix(<area>): <what was fixed>   ← bug fix
doc: <what was documented>      ← documentation
```

Examples from the repo:
```
variable: reject offsets that overflow uint32 bounds
fix(link): add missing BPF_F_REPLACE flag for RawAttachProgram
doc: clarify Address vs Offset on Uprobe/Uretprobe
CODEOWNERS: allow reviewers to merge docs/
```

## Running Tests

```bash
# needs sudo for BPF privileges — run in Lima VM:
go test -exec sudo ./...

# test against a specific kernel version (needs vimto):
vimto -- go test ./...
vimto -kernel :mainline -- go test ./...
```

## Project Roles (contribution ladder)

```
contributor  → Triage role, may be asked to review/help
reviewer     → Write role, CODEOWNER of part of codebase
maintainer   → Admin role, manages releases, merges PRs
```

Start by getting one PR merged → maintainers can add you to the team.

## Regenerating test data

```bash
make          # requires Docker
# or with Podman:
make CONTAINER_ENGINE=podman CONTAINER_RUN_ARGS=
```

---

## Potential Contribution Areas
> Add here when you find something during learning

- [ ] Issue #506: https://github.com/cilium/ebpf/issues/506 — good first issue, still open as of 2026-06-19, one person expressed interest but no PR yet. Comment first before starting.

## Useful Links
- GitHub issues: https://github.com/cilium/ebpf/issues
- Good first issues: https://github.com/cilium/ebpf/issues?q=is%3Aopen+label%3A%22good+first+issue%22
- Cilium Slack: https://ebpf.io/slack → #ebpf-go-dev channel
- DCO: https://developercertificate.org/
- Contributing guide: https://ebpf-go.dev/contributing/
