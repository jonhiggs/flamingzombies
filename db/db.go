package db

import "fmt"

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

var LockCh = make(chan Lock)
var StateCh = make(chan State)

func Start() {
	go func() {
		select {
		case l := <-LockCh:
			if l.Locked {
				fmt.Println("Locking ", l.Hash)
				saveLock(l.Hash)
			} else {
				fmt.Println("Unlocking ", l.Hash)
				deleteLock(l.Hash)
			}
		}
	}()
}

func saveLock(hash uint32) {
	deleteLock(hash)
	taskLocks = append(taskLocks, hash)
}

func deleteLock(hash uint32) {
	newLocks := []uint32{}
	for _, h := range taskLocks {
		if h != hash {
			newLocks = append(newLocks, h)
		}
	}

	taskLocks = newLocks
}
