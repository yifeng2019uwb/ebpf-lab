package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

// map keys — matches probe.bpf.c index comments
const (
	keyKprobe    = uint32(0)
	keyKretprobe = uint32(1)
	keyUprobe    = uint32(2)
	keyUretprobe = uint32(3)
)

// result mirrors the BPF struct event (40 bytes):
//
//	offset  0  pid    uint32
//	offset  4  family uint16  kprobe/kretprobe: socket family (AF_INET=2, AF_INET6=10)
//	offset  6  sport  uint16  kprobe/kretprobe: source port
//	offset  8  dport  uint16  kprobe/kretprobe: dest port
//	offset 10  fd     uint16  uprobe: write() fd (2=stderr at entry, garbage at return)
//	offset 12  count  uint32  uprobe: write() byte count (valid at entry, garbage at return)
//	offset 16  retval int64   uretprobe: bytes written (garbage if wrong hook)
//	offset 24  probe_type[16] (not read in Go)
type result struct {
	attachErr error
	pid       uint32
	family    uint16
	sport     uint16
	dport     uint16
	fd        uint16
	count     uint32
	retval    int64
	fired     bool
}

func (r result) String() string {
	if r.attachErr != nil {
		return fmt.Sprintf("attach error: %v", r.attachErr)
	}
	if !r.fired {
		return "program did not fire (no event)"
	}
	return fmt.Sprintf("pid=%-6d family=%-6d sport=%-6d dport=%-6d fd=%-4d count=%-6d retval=%d",
		r.pid, r.family, r.sport, r.dport, r.fd, r.count, r.retval)
}

func main() {
	if os.Getuid() != 0 {
		log.Fatal("must run as root")
	}
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatalf("remove memlock: %v", err)
	}

	objs := probeObjects{}
	if err := loadProbeObjects(&objs, nil); err != nil {
		log.Fatalf("load BPF objects: %v", err)
	}
	defer objs.Close()

	libcPath := findLibc()
	if libcPath == "" {
		log.Fatal("could not find libc")
	}
	ex, err := link.OpenExecutable(libcPath)
	if err != nil {
		log.Fatalf("open libc: %v", err)
	}

	fmt.Println("══════════════════════════════════════════════════════════")
	fmt.Println(" Issue #1497 — full test matrix")
	fmt.Println(" Goal: show library returns no error for wrong combos.")
	fmt.Println(" Each pair: CORRECT vs WRONG — compare values directly.")
	fmt.Println("══════════════════════════════════════════════════════════")

	// ── Pair 1: kprobe program ────────────────────────────────────────────────
	// kprobe_accept reads PT_REGS_RC (rax): garbage at entry, real sock* at return.
	fmt.Println("\n── Pair 1: kprobe program (key 0) ──")
	r1c := runTest(&objs, func() (link.Link, error) {
		return link.Kprobe("inet_csk_accept", objs.KprobeAccept, nil)
	}, triggerTCP, keyKprobe)
	r1w := runTest(&objs, func() (link.Link, error) {
		return link.Kretprobe("inet_csk_accept", objs.KprobeAccept, nil)
	}, triggerTCP, keyKprobe)
	compare("kprobe prog + Kprobe()    [CORRECT]", r1c,
		"kprobe prog + Kretprobe() [WRONG]  ", r1w)

	// ── Pair 2: kretprobe program ─────────────────────────────────────────────
	// Issue #1490 exact case: kretprobe prog via Kprobe() → garbage family/sport/dport.
	fmt.Println("\n── Pair 2: kretprobe program (key 1) ──")
	r2c := runTest(&objs, func() (link.Link, error) {
		return link.Kretprobe("inet_csk_accept", objs.KretprobeAccept, nil)
	}, triggerTCP, keyKretprobe)
	r2w := runTest(&objs, func() (link.Link, error) {
		return link.Kprobe("inet_csk_accept", objs.KretprobeAccept, nil)
	}, triggerTCP, keyKretprobe)
	compare("kretprobe prog + Kretprobe() [CORRECT]", r2c,
		"kretprobe prog + Kprobe()    [WRONG]  ", r2w)

	// ── Pair 3: uprobe program ────────────────────────────────────────────────
	// uprobe_write captures fd and count (separate fields, not reusing sport/dport).
	// CORRECT (entry):  fd=2 (stderr), count=8 ("trigger\n")
	// WRONG (return):   argument registers clobbered → garbage fd/count
	fmt.Println("\n── Pair 3: uprobe program (key 2) ──")
	fmt.Println("  (fd=2=stderr, count=8 at entry; registers clobbered at return → garbage)")
	r3c := runTest(&objs, func() (link.Link, error) {
		return ex.Uprobe("write", objs.UprobeWrite, nil)
	}, triggerWrite, keyUprobe)
	r3w := runTest(&objs, func() (link.Link, error) {
		return ex.Uretprobe("write", objs.UprobeWrite, nil)
	}, triggerWrite, keyUprobe)
	compare("uprobe prog + Uprobe()    [CORRECT]", r3c,
		"uprobe prog + Uretprobe() [WRONG]  ", r3w)

	// ── Pair 4: uretprobe program ─────────────────────────────────────────────
	// CORRECT (return): retval = actual bytes written
	// WRONG (entry):    rax not set at entry → garbage retval
	fmt.Println("\n── Pair 4: uretprobe program (key 3) ──")
	r4c := runTest(&objs, func() (link.Link, error) {
		return ex.Uretprobe("write", objs.UretprobeWrite, nil)
	}, triggerWrite, keyUretprobe)
	r4w := runTest(&objs, func() (link.Link, error) {
		return ex.Uprobe("write", objs.UretprobeWrite, nil)
	}, triggerWrite, keyUretprobe)
	compare("uretprobe prog + Uretprobe() [CORRECT]", r4c,
		"uretprobe prog + Uprobe()    [WRONG]  ", r4w)

	// ── Cross-domain A: kprobe/kretprobe programs on write() userspace hooks ──
	// kprobe_accept reads rax as sock* — at write() return rax=bytes_written,
	// which is interpreted as sock ptr → BPF_CORE_READ safely returns 0.
	fmt.Println("\n── Cross-domain A: kernel programs on write() hooks (key 0 and key 1) ──")

	rca1 := runTest(&objs, func() (link.Link, error) {
		return ex.Uprobe("write", objs.KprobeAccept, nil)
	}, triggerWrite, keyKprobe)
	fmt.Printf("  kprobe prog   + Uprobe(write)     [cross]: %s\n", rca1)

	rca2 := runTest(&objs, func() (link.Link, error) {
		return ex.Uretprobe("write", objs.KprobeAccept, nil)
	}, triggerWrite, keyKprobe)
	fmt.Printf("  kprobe prog   + Uretprobe(write)  [cross]: %s\n", rca2)

	rca3 := runTest(&objs, func() (link.Link, error) {
		return ex.Uprobe("write", objs.KretprobeAccept, nil)
	}, triggerWrite, keyKretprobe)
	fmt.Printf("  kretprobe prog + Uprobe(write)    [cross]: %s\n", rca3)

	rca4 := runTest(&objs, func() (link.Link, error) {
		return ex.Uretprobe("write", objs.KretprobeAccept, nil)
	}, triggerWrite, keyKretprobe)
	fmt.Printf("  kretprobe prog + Uretprobe(write) [cross]: %s\n", rca4)

	// ── Cross-domain B: uprobe/uretprobe programs on inet_csk_accept kernel hooks ──
	// uretprobe_write reads rax as retval — at kretprobe return rax=sock_ptr
	// (a kernel address like 0xffff888...) → shows as huge retval value.
	fmt.Println("\n── Cross-domain B: userspace programs on inet_csk_accept hooks (key 2 and key 3) ──")

	rcb1 := runTest(&objs, func() (link.Link, error) {
		return link.Kprobe("inet_csk_accept", objs.UprobeWrite, nil)
	}, triggerTCP, keyUprobe)
	fmt.Printf("  uprobe prog    + Kprobe(accept)    [cross]: %s\n", rcb1)

	rcb2 := runTest(&objs, func() (link.Link, error) {
		return link.Kretprobe("inet_csk_accept", objs.UprobeWrite, nil)
	}, triggerTCP, keyUprobe)
	fmt.Printf("  uprobe prog    + Kretprobe(accept)  [cross]: %s\n", rcb2)

	rcb3 := runTest(&objs, func() (link.Link, error) {
		return link.Kprobe("inet_csk_accept", objs.UretprobeWrite, nil)
	}, triggerTCP, keyUretprobe)
	fmt.Printf("  uretprobe prog + Kprobe(accept)    [cross]: %s\n", rcb3)

	rcb4 := runTest(&objs, func() (link.Link, error) {
		return link.Kretprobe("inet_csk_accept", objs.UretprobeWrite, nil)
	}, triggerTCP, keyUretprobe)
	fmt.Printf("  uretprobe prog + Kretprobe(accept)  [cross]: %s\n", rcb4)
	fmt.Println("  Note: uretprobe prog at kretprobe return reads sock_ptr as retval → huge number")

	// ── Pin/FD Test: sectionName == "" case ────────────────────────────────────
	// When programs are loaded via pin/FD, sectionName is empty.
	// Validation is skipped, but warning is printed to stderr.
	// Caller is responsible for using correct link function.
	fmt.Println("\n── Pin/FD Test: Programs loaded via pin (sectionName == \"\") ──")
	fmt.Println("  (validation skipped, warning printed to stderr)")

	// Create temp directory for pinned programs
	tmpDir := "/tmp/ebpf_pin_test"
	_ = os.RemoveAll(tmpDir)
	if err := os.Mkdir(tmpDir, 0755); err != nil {
		log.Fatalf("mkdir %s: %v", tmpDir, err)
	}
	defer os.RemoveAll(tmpDir)

	// Pin all four programs
	kprobePinPath := tmpDir + "/kprobe.o"
	kretprobePinPath := tmpDir + "/kretprobe.o"
	uprobePinPath := tmpDir + "/uprobe.o"
	uretprobePinPath := tmpDir + "/uretprobe.o"

	if err := objs.KprobeAccept.Pin(kprobePinPath); err != nil {
		log.Fatalf("pin kprobe: %v", err)
	}
	if err := objs.KretprobeAccept.Pin(kretprobePinPath); err != nil {
		log.Fatalf("pin kretprobe: %v", err)
	}
	if err := objs.UprobeWrite.Pin(uprobePinPath); err != nil {
		log.Fatalf("pin uprobe: %v", err)
	}
	if err := objs.UretprobeWrite.Pin(uretprobePinPath); err != nil {
		log.Fatalf("pin uretprobe: %v", err)
	}

	// Load pinned programs (sectionName will be "")
	pinnedKprobe, err := ebpf.LoadPinnedProgram(kprobePinPath, nil)
	if err != nil {
		log.Fatalf("load pinned kprobe: %v", err)
	}
	defer pinnedKprobe.Close()

	pinnedKretprobe, err := ebpf.LoadPinnedProgram(kretprobePinPath, nil)
	if err != nil {
		log.Fatalf("load pinned kretprobe: %v", err)
	}
	defer pinnedKretprobe.Close()

	pinnedUprobe, err := ebpf.LoadPinnedProgram(uprobePinPath, nil)
	if err != nil {
		log.Fatalf("load pinned uprobe: %v", err)
	}
	defer pinnedUprobe.Close()

	pinnedUretprobe, err := ebpf.LoadPinnedProgram(uretprobePinPath, nil)
	if err != nil {
		log.Fatalf("load pinned uretprobe: %v", err)
	}
	defer pinnedUretprobe.Close()

	// Test pinned programs with correct and wrong attachments
	fmt.Println("\n  Pinned kprobe (sectionName=\"\"):")
	rpkc := runTest(&objs, func() (link.Link, error) {
		return link.Kprobe("inet_csk_accept", pinnedKprobe, nil)
	}, triggerTCP, keyKprobe)
	fmt.Printf("    + Kprobe()    [CORRECT]: %s\n", rpkc)

	rpkw := runTest(&objs, func() (link.Link, error) {
		return link.Kretprobe("inet_csk_accept", pinnedKprobe, nil)
	}, triggerTCP, keyKprobe)
	fmt.Printf("    + Kretprobe() [WRONG]:   %s\n", rpkw)

	fmt.Println("\n  Pinned kretprobe (sectionName=\"\"):")
	rprc := runTest(&objs, func() (link.Link, error) {
		return link.Kretprobe("inet_csk_accept", pinnedKretprobe, nil)
	}, triggerTCP, keyKretprobe)
	fmt.Printf("    + Kretprobe() [CORRECT]: %s\n", rprc)

	rprw := runTest(&objs, func() (link.Link, error) {
		return link.Kprobe("inet_csk_accept", pinnedKretprobe, nil)
	}, triggerTCP, keyKretprobe)
	fmt.Printf("    + Kprobe()    [WRONG]:   %s\n", rprw)

	fmt.Println("\n  Pinned uprobe (sectionName=\"\"):")
	rpuc := runTest(&objs, func() (link.Link, error) {
		return ex.Uprobe("write", pinnedUprobe, nil)
	}, triggerWrite, keyUprobe)
	fmt.Printf("    + Uprobe()    [CORRECT]: %s\n", rpuc)

	rpuw := runTest(&objs, func() (link.Link, error) {
		return ex.Uretprobe("write", pinnedUprobe, nil)
	}, triggerWrite, keyUprobe)
	fmt.Printf("    + Uretprobe() [WRONG]:   %s\n", rpuw)

	fmt.Println("\n  Pinned uretprobe (sectionName=\"\"):")
	rpurc := runTest(&objs, func() (link.Link, error) {
		return ex.Uretprobe("write", pinnedUretprobe, nil)
	}, triggerWrite, keyUretprobe)
	fmt.Printf("    + Uretprobe() [CORRECT]: %s\n", rpurc)

	rpurw := runTest(&objs, func() (link.Link, error) {
		return ex.Uprobe("write", pinnedUretprobe, nil)
	}, triggerWrite, keyUretprobe)
	fmt.Printf("    + Uprobe()    [WRONG]:   %s\n", rpurw)

	fmt.Println("\n  Note: Pinned programs have sectionName=\"\" → validation skipped")
	fmt.Println("        Warning printed to stderr, but attachment behavior depends on kernel.")

	fmt.Println("\n══════════════════════════════════════════════════════════")
	fmt.Println(" Phase 1 done. WRONG pairs show garbage vs correct values.")
	fmt.Println(" Phase 2: modify cilium/ebpf → WRONG cases return errors.")
	fmt.Println("══════════════════════════════════════════════════════════")
}

// compare prints two results side by side so garbage is obvious.
func compare(labelC string, correct result, labelW string, wrong result) {
	fmt.Printf("  %-45s %s\n", labelC, correct)
	fmt.Printf("  %-45s %s\n", labelW, wrong)
	if correct.attachErr == nil && wrong.attachErr == nil && correct.fired && wrong.fired {
		if correct.family != wrong.family || correct.sport != wrong.sport || correct.dport != wrong.dport ||
			correct.fd != wrong.fd || correct.count != wrong.count || correct.retval != wrong.retval {
			fmt.Println("  → values differ: WRONG is garbage (proves bug)")
		} else {
			fmt.Println("  → values same: may be coincidence or uninitialised zero")
		}
	}
}

// runTest attaches, triggers, reads map entry, cleans up. Returns result.
func runTest(objs *probeObjects, attach func() (link.Link, error), trigger func(), key uint32) result {
	clearMapEntry(objs, key)

	lnk, err := attach()
	if err != nil {
		return result{attachErr: err}
	}
	defer lnk.Close()

	trigger()
	return readMapEntry(objs, key)
}

func readMapEntry(objs *probeObjects, key uint32) result {
	var raw [40]byte
	if err := objs.Results.Lookup(key, &raw); err != nil {
		return result{attachErr: fmt.Errorf("map lookup: %w", err)}
	}
	r := result{
		pid:    binary.LittleEndian.Uint32(raw[0:4]),
		family: binary.LittleEndian.Uint16(raw[4:6]),
		sport:  binary.LittleEndian.Uint16(raw[6:8]),
		dport:  binary.LittleEndian.Uint16(raw[8:10]),
		fd:     binary.LittleEndian.Uint16(raw[10:12]),
		count:  binary.LittleEndian.Uint32(raw[12:16]),
		retval: int64(binary.LittleEndian.Uint64(raw[16:24])),
	}
	r.fired = r.pid != 0
	return r
}

func clearMapEntry(objs *probeObjects, key uint32) {
	var zero [40]byte
	_ = objs.Results.Put(key, zero)
}

func triggerTCP() {
	time.Sleep(100 * time.Millisecond)
	conn, err := net.DialTimeout("tcp", "localhost:22", time.Second)
	if err == nil {
		conn.Close()
	}
	time.Sleep(100 * time.Millisecond)
}

func triggerWrite() {
	time.Sleep(100 * time.Millisecond)
	fmt.Fprint(os.Stderr, "trigger\n")
	time.Sleep(100 * time.Millisecond)
}

func findLibc() string {
	for _, p := range []string{
		"/lib/x86_64-linux-gnu/libc.so.6",
		"/lib/aarch64-linux-gnu/libc.so.6",
		"/lib64/libc.so.6",
		"/usr/lib/libc.so.6",
	} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
