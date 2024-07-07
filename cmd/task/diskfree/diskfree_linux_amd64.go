package main

import "golang.org/x/sys/unix"

func diskfree(dir string) {
	var stat unix.Statfs_t
	unix.Statfs(dir, &stat)

	// Available blocks * size per block = available space in bytes, then
	// converted to KB
	var free uint64
	free = (stat.Bavail * uint64(stat.Bsize)) / 1024

	bytes <- int64(free)
}
