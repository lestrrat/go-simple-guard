package guard

import "sync"

// Guard is the interface for all guards. Most of your code should
// probably specify this as either return type or variable type.
type Guard interface {
	Fire() error
	Cancel() error
}

type guardFiredErr struct{}
type guardCanceledErr struct{}

var (
	errFired    = guardFiredErr{}
	errCanceled = guardCanceledErr{}
)

func (_ guardFiredErr) Fired() bool {
	return true
}

func (_ guardFiredErr) Error() string {
	return "guard has already been fired"
}

func (_ guardCanceledErr) Canceled() bool {
	return true
}

func (_ guardCanceledErr) Error() string {
	return "guard has already been canceled"
}

// Nil is a special guard that does nothing. Use it in tests or
// when you just need to pass a dummy guard to fulfill some function call.
var Nil nilGuard

type nilGuard struct{}

const (
	stFired    = 0x001
	stCanceled = 0x010
)

// CB is the most generic guard type, one that executes the given callback.
type CB struct {
	mutex  sync.Mutex
	state  int8
	onFire func() error
}
