package main

const (
	CTL_VFS        = 10
	VFS_GENERIC    = 0
	VFS_BCACHESTAT = 3
)

func swapfree() {
	//uvmexpb, err := unix.SysctlRaw("vm.uvmexp")
	//if err != nil {
	//	panic(err)
	//}

	//mib := [3]_C_int{CTL_VFS, VFS_GENERIC, VFS_BCACHESTAT}
	//bcstatsb, err := sysctl(mib[:])
	//if err != nil {
	//	panic(err)
	//}

	//uvmexp := *(*unix.Uvmexp)(unsafe.Pointer(&uvmexpb[0]))

	//bytes <- int64(uvmexp.Active)
}
