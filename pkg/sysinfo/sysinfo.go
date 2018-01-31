package sysinfo

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	humanize "github.com/dustin/go-humanize"
)

// SysInfo contains system statistics
type SysInfo map[string]string

// GetSysInfo returns useful system statistics.
func GetSysInfo() SysInfo {
	host, err := os.Hostname()
	if err != nil {
		host = ""
	}
	memstats := &runtime.MemStats{}
	runtime.ReadMemStats(memstats)
	return SysInfo{
		"host.name":      host,
		"host.os":        runtime.GOOS,
		"host.arch":      runtime.GOARCH,
		"host.lang":      runtime.Version(),
		"host.cpus":      strconv.Itoa(runtime.NumCPU()),
		"mem.used":       humanize.Bytes(memstats.Alloc),
		"mem.total":      humanize.Bytes(memstats.Sys),
		"mem.heap.used":  humanize.Bytes(memstats.HeapAlloc),
		"mem.heap.total": humanize.Bytes(memstats.HeapSys),
	}
}

// String implements fmt.Stringer interface.
func (s SysInfo) String() string {

	str := ""
	str += "Host:" + s["host.name"] + " | "
	str += "OS:" + s["host.os"] + " | "
	str += "Arch:" + s["host.arch"] + " | "
	str += "Lang:" + s["host.lang"] + " | "
	str += "CPUs:" + s["host.cpus"] + " | "
	str += "Mem:" + s["mem.used"] + "/" + s["mem.total"] + " | "
	str += "Heap:" + s["mem.heap.used"] + "/" + s["mem.heap.total"]

	return str
}

// Check the interfaces are satisfied
var (
	_ fmt.Stringer = &SysInfo{}
)
