package main

import "golang.org/x/sys/unix"

func swapfree() {
	var s unix.Sysinfo_t
	unix.Sysinfo(&s)

	var free uint64
	free = s.Freeswap / 1024

	bytes <- int64(free)
}
