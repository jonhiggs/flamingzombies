package fz

type State int8

const (
	STATE_UNKNOWN = State(iota)
	STATE_FAIL
	STATE_OK
)

func (s State) String() string {
	switch s {
	case STATE_FAIL:
		return "fail"
	case STATE_OK:
		return "ok"
	default:
		return "unknown"
	}
}
