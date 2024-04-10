package fz

import "fmt"

type StateRecord struct {
	Hash   uint32
	Status bool
}

type State struct {
	Hash    uint32
	Retries int
	History uint32
}

var States []State
var StateRecordCh = make(chan StateRecord, 100)

func RecordStates() {
	go func() {
		for {
			select {
			case r := <-StateRecordCh:
				st := FindState(r.Hash)
				statusA := st.Status()
				st.Append(r.Status)
				st = FindState(r.Hash)
				statusB := st.Status()

				fmt.Printf("recording status of %d: %b: %b -> %b\n", st.Hash, st.History, statusA, statusB)

				if statusA != statusB {
					// TODO: signal the notification channel
					fmt.Printf(
						"status of %d has changed from %d to %d\n",
						st.Hash,
						statusA,
						statusB,
					)
				}
			}
		}
	}()
}

func FindState(hash uint32) *State {
	for i, st := range States {
		if st.Hash == hash {
			return &States[i]
		}
	}

	panic("state could not be found")
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
