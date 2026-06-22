# Unit Tests for prog.go sectionName Field (Issue #1497)

## Changes Made to prog.go
1. Added `sectionName string` field to Program struct
2. Populated field in NewProgramWithOptions() from spec.SectionName
3. Added SectionName() getter method

## Unit Tests Required

### 1. Test Program.SectionName() Method

```go
func TestProgramSectionName(t *testing.T) {
    // Test: Program created with sectionName should return it
    spec := &ebpf.ProgramSpec{
        Type: ebpf.SocketFilter,
        SectionName: "kprobe/test_symbol",
        License: "MIT",
        Instructions: basicInstructions,
    }
    
    prog, err := ebpf.NewProgram(spec)
    if err != nil {
        t.Fatal(err)
    }
    defer prog.Close()
    
    // Verify SectionName() returns the value
    if got := prog.SectionName(); got != "kprobe/test_symbol" {
        t.Fatalf("SectionName() = %q, want %q", got, "kprobe/test_symbol")
    }
}
```

### 2. Test Different Section Names

```go
func TestProgramSectionNames(t *testing.T) {
    testCases := []string{
        "kprobe/inet_csk_accept",
        "kretprobe/inet_csk_accept",
        "uprobe/write",
        "uretprobe/write",
        "tracepoint/syscalls/sys_enter_open",
        "",  // empty (pin-loaded programs)
    }
    
    for _, sectionName := range testCases {
        spec := &ebpf.ProgramSpec{
            Type: ebpf.Kprobe,
            SectionName: sectionName,
            License: "MIT",
            Instructions: basicInstructions,
        }
        
        prog, err := ebpf.NewProgram(spec)
        if err != nil {
            t.Fatal(err)
        }
        
        if got := prog.SectionName(); got != sectionName {
            t.Fatalf("SectionName() = %q, want %q", got, sectionName)
        }
        prog.Close()
    }
}
```

### 3. Test Program.Clone() Preserves sectionName

```go
func TestProgramCloneSectionName(t *testing.T) {
    spec := &ebpf.ProgramSpec{
        Type: ebpf.SocketFilter,
        SectionName: "kprobe/test",
        License: "MIT",
        Instructions: basicInstructions,
    }
    
    orig, err := ebpf.NewProgram(spec)
    if err != nil {
        t.Fatal(err)
    }
    defer orig.Close()
    
    cloned, err := orig.Clone()
    if err != nil {
        t.Fatal(err)
    }
    defer cloned.Close()
    
    // Verify sectionName is preserved on clone
    if got := cloned.SectionName(); got != orig.SectionName() {
        t.Fatalf("cloned SectionName() = %q, want %q", got, orig.SectionName())
    }
}
```

### 4. Test NewProgramFromFD Has Empty sectionName

```go
func TestNewProgramFromFDSectionName(t *testing.T) {
    // Programs loaded from FD should have empty sectionName
    // (This is typically tested with pin/FD loads)
    
    // Note: This test is harder to write without actual FD
    // but the behavior is: LoadPinnedProgram → sectionName = ""
}
```

### 5. Test Program Struct Initialization

```go
func TestProgramStructSectionName(t *testing.T) {
    // Verify all Program constructor paths properly set sectionName
    
    // Path 1: NewProgram (with sectionName)
    spec1 := &ebpf.ProgramSpec{
        Type: ebpf.SocketFilter,
        SectionName: "test/symbol",
        License: "MIT",
        Instructions: basicInstructions,
    }
    prog1, _ := ebpf.NewProgram(spec1)
    defer prog1.Close()
    if prog1.SectionName() != "test/symbol" {
        t.Fatal("NewProgram: sectionName not set")
    }
    
    // Path 2: NewProgram (without sectionName - default empty)
    spec2 := &ebpf.ProgramSpec{
        Type: ebpf.SocketFilter,
        License: "MIT",
        Instructions: basicInstructions,
        // SectionName not set - should default to ""
    }
    prog2, _ := ebpf.NewProgram(spec2)
    defer prog2.Close()
    if prog2.SectionName() != "" {
        t.Fatal("NewProgram: default sectionName should be empty")
    }
}
```

## Test File Location

Add these tests to: `/Users/yifengzh/workspace/ebpf/prog_test.go`

## Integration with Link Tests

These prog.go tests verify the struct changes.
The link tests (kprobe_test.go, uprobe_test.go) verify the validation logic uses sectionName correctly.

## Pre-existing Test Failures

The following map test failures are NOT related to our changes:
- TestMapBatch/PerCPUHash
- TestPerfEventArrayCompatible

These should be investigated separately and fixed before committing.

## Checklist

- [ ] Add TestProgramSectionName
- [ ] Add TestProgramSectionNames
- [ ] Add TestProgramCloneSectionName
- [ ] Add TestProgramStructSectionName
- [ ] Run: `go test ./... -run TestProgram*`
- [ ] Verify all 4 tests pass
- [ ] Document any failures unrelated to sectionName
