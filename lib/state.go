package fz

// the returned values are:
//
//	-1: unknown
//	 0: down
//	 1: up
//func (st State) Status() int {
//	var mask uint32
//	for i := 0; i < st.Retries; i++ {
//		mask = mask << 1
//		mask += 1
//	}
//	v := st.History & mask
//
//	if v == 0 {
//		return 0
//	}
//
//	if v == mask {
//		return 1
//	}
//
//	return -1
//}
