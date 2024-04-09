package db

type Lock struct {
	Hash   uint32
	Locked bool
}

type State struct {
	Hash  uint32
	State int
}

var taskStates []int
var taskLocks []uint32

var LockCh = make(chan Lock, 100)
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
