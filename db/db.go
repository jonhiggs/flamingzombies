package db

type Lock struct {
	Hash   uint32
	Locked bool
}

var taskLocks []uint32
var LockCh = make(chan Lock, 100)

type State struct {
	Hash  uint32
	State uint32
}

var states []State
var StateCh = make(chan State, 100)

func Start() {
	go func() {
		for {
			select {
			case l := <-LockCh:
				if l.Locked {
					lock(l.Hash)
				} else {
					unlock(l.Hash)
				}
			case s := <-StateCh:
				appendState(s.Hash, s.State)
			}
		}
	}()
}

func lock(hash uint32) bool {
	unlock(hash)
	taskLocks = append(taskLocks, hash)
	return true
}

func unlock(hash uint32) bool {
	newLocks := []uint32{}
	for _, h := range taskLocks {
		if h != hash {
			newLocks = append(newLocks, h)
		}
	}

	taskLocks = newLocks
	return true
}

func IsLocked(hash uint32) bool {
	for _, h := range taskLocks {
		if h == hash {
			return true
		}
	}

	return false
}

func appendState(hash uint32, state uint32) {
	//st := curState(hash)
	//st = st << 1
	//if state > 0 {
	//	st += 1
	//}

	//if curState == nil {
	//	curState = st
	//} else {
	//	curState.State = curState.State << 1
	//	if st.State < 0 {
	//		curState.State += 1
	//	}
	//}
}

//func currentState(hash uint32) uint32 {
//	for _, s := range states {
//		if s.Hash() == hash {
//			return s.State
//		}
//	}
//	return 0
//}
