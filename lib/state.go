package fz

import "fmt"

type State struct {
	Hash    uint32
	Retries int
	History uint32
}

var states []State
var StateCh = make(chan State, 100)

func RecordStates() {
	go func() {
		for {
			select {
			case s := <-StateCh:
				fmt.Println(s)
			}
		}
	}()
}

func (st *State) Append(b bool) {
	st.History = st.History << 1
	if b {
		st.History += 1
	}
}

// the returned values are:
//
//	-1: unknown
//	 0: down
//	 1: up

func (st State) Status() int {
	var mask uint32
	for i := 0; i < st.Retries; i++ {
		mask = mask << 1
		mask += 1
	}
	v := st.History & mask

	if v == 0 {
		return 0
	}

	if v == mask {
		return 1
	}

	return -1
}
