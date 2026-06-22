# Unit Test Plan for sectionName Validation (Issue #1497)

## Problem
- Existing tests print warnings because they load programs via pin/FD (sectionName == "")
- New validation code triggers warnings for these pin-loaded programs
- Tests still pass, but output is cluttered

## Solution

### 1. New Unit Tests for Validation (kprobe_test.go)

Add tests that explicitly test validation with ELF-loaded programs:

```go
func TestKprobeValidation(t *testing.T) {
    // Test: kprobe section with Kprobe() should succeed
    // Test: kprobe section with Kretprobe() should error
    // Test: kretprobe section with Kprobe() should error
    // Test: kretprobe section with Kretprobe() should succeed
}

func TestKprobeWrongHookType(t *testing.T) {
    // Create program with "kretprobe/" section name
    // Try to attach via Kprobe()
    // Expect error: "program is kretprobe/..., cannot attach via Kprobe()"
}

func TestCrossDomainValidation(t *testing.T) {
    // Test uprobe section with Kprobe() → error
    // Test kprobe section with Uprobe() → error
}
```

### 2. Suppress Warnings in Existing Tests

Redirect stderr during tests to avoid cluttering output.

### 3. Create Test Helpers

Add helper to create programs with specific sectionName values:

```go
// In helpers_test.go
func loadProgramWithSection(tb testing.TB, progType ebpf.ProgramType, sectionName string) *ebpf.Program {
    spec := &ebpf.ProgramSpec{
        Type: progType,
        SectionName: sectionName,
        License: "MIT",
        Instructions: asm.Instructions{
            asm.Mov.Imm(asm.R0, 0),
            asm.Return(),
        },
    }
    
    prog, err := ebpf.NewProgram(spec)
    if err != nil {
        tb.Fatal(err)
    }
    
    tb.Cleanup(func() {
        prog.Close()
    })
    
    return prog
}
```

## Test Cases to Add

### kprobe_test.go
- ✓ kprobe section + Kprobe() = SUCCESS
- ✓ kprobe section + Kretprobe() = ERROR
- ✓ kretprobe section + Kretprobe() = SUCCESS
- ✓ kretprobe section + Kprobe() = ERROR
- ✓ uprobe section + Kprobe() = ERROR (cross-domain)
- ✓ uretprobe section + Kretprobe() = ERROR (cross-domain)

### uprobe_test.go
- ✓ uprobe section + Uprobe() = SUCCESS
- ✓ uprobe section + Uretprobe() = ERROR
- ✓ uretprobe section + Uretprobe() = SUCCESS
- ✓ uretprobe section + Uprobe() = ERROR
- ✓ kprobe section + Uprobe() = ERROR (cross-domain)
- ✓ kretprobe section + Uretprobe() = ERROR (cross-domain)

## Implementation Order

1. Add helpers to create programs with specific sectionName
2. Add new validation test functions
3. Update existing tests to suppress warnings (if needed)
4. Verify all tests pass cleanly

## Expected Results
- 12 new validation tests covering all section/hook combinations
- Existing tests pass without warnings
- Full coverage of issue #1497 validation logic
