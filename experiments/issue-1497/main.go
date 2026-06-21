package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"time"

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

type result struct {
	attachErr error
	pid       uint32
	family    uint16
	sport     uint16
	dport     uint16
	retval    int64
	fired     bool // true if program wrote to map (pid != 0)
}

func (r result) String() string {
	if r.attachErr != nil {
		return fmt.Sprintf("attach error: %v", r.attachErr)
	}
	if !r.fired {
		return "program did not fire (no event)"
	}
	return fmt.Sprintf("pid=%-6d family=%-6d sport=%-6d dport=%-6d retval=%d",
		r.pid, r.family, r.sport, r.dport, r.retval)
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
	// kprobe_accept reads PT_REGS_RC at entry (wrong) vs return (right)
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
	// kretprobe_accept — this is issue #1490 exact case
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
	// uprobe_write only captures pid — same value at entry and return.
	// Values won't differ between CORRECT and WRONG here.
	// The proof of bug is: no attach error returned for the wrong combo.
	fmt.Println("\n── Pair 3: uprobe program (key 2) ──")
	fmt.Println("  (uprobe_write captures only pid; values won't differ — bug proven by no attach error)")
	r3c := runTest(&objs, func() (link.Link, error) {
		return ex.Uprobe("write", objs.UprobeWrite, nil)
	}, triggerWrite, keyUprobe)
	r3w := runTest(&objs, func() (link.Link, error) {
		return ex.Uretprobe("write", objs.UprobeWrite, nil)
	}, triggerWrite, keyUprobe)
	compare("uprobe prog + Uprobe()    [CORRECT]", r3c,
		"uprobe prog + Uretprobe() [WRONG]  ", r3w)

	// ── Pair 4: uretprobe program ─────────────────────────────────────────────
	fmt.Println("\n── Pair 4: uretprobe program (key 3) ──")
	r4c := runTest(&objs, func() (link.Link, error) {
		return ex.Uretprobe("write", objs.UretprobeWrite, nil)
	}, triggerWrite, keyUretprobe)
	r4w := runTest(&objs, func() (link.Link, error) {
		return ex.Uprobe("write", objs.UretprobeWrite, nil)
	}, triggerWrite, keyUretprobe)
	compare("uretprobe prog + Uretprobe() [CORRECT]", r4c,
		"uretprobe prog + Uprobe()    [WRONG]  ", r4w)

	// ── Cross-domain: kernel prog attached to userspace hook ──────────────────
	// All share prog.Type()==Kprobe so library validation passes.
	// Kernel may reject or silently accept — result is informative either way.
	fmt.Println("\n── Cross-domain (key 0 and key 2) ──")
	r5 := runTest(&objs, func() (link.Link, error) {
		return ex.Uprobe("write", objs.KprobeAccept, nil)
	}, triggerWrite, keyKprobe)
	fmt.Printf("  kprobe prog + Uprobe() [cross-domain]: %s\n", r5)

	r6 := runTest(&objs, func() (link.Link, error) {
		return link.Kprobe("inet_csk_accept", objs.UprobeWrite, nil)
	}, triggerTCP, keyUprobe)
	fmt.Printf("  uprobe prog + Kprobe() [cross-domain]: %s\n", r6)

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
		if correct.family != wrong.family || correct.sport != wrong.sport || correct.retval != wrong.retval {
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
	var raw [40]byte // sizeof(struct event) = 40; cilium/ebpf checks size matches exactly
	if err := objs.Results.Lookup(key, &raw); err != nil {
		return result{attachErr: fmt.Errorf("map lookup: %w", err)}
	}
	r := result{
		pid:    binary.LittleEndian.Uint32(raw[0:4]),
		family: binary.LittleEndian.Uint16(raw[4:6]),
		sport:  binary.LittleEndian.Uint16(raw[6:8]),
		dport:  binary.LittleEndian.Uint16(raw[8:10]),
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
