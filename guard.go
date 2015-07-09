package guard

/*
package guard provides a simple construct to help you write a RAII-like
construct in Go.

Go doesn't provide a deterministic way to fire code at garbage collection
time, but you can sort of do it when `defer` gets fired.

	func Foo() {
		defer CleanupCode()

		...
	}

The guard package gives you one more additional layer of functionality.
For example, if you're doing a database operation, you might want to
register a `Rollback()` call, only to make sure that in case you return
before committing, you make sure your previous operations are discarded:

	func DatabaseOperation(db *sql.DB) {
		tx := db.Begin()
		defer tx.Rollback()

		... // database operation that may fail
		if err != nil {
			return
		}

		tx.Commit()
	}

Except, if the operation is successful, you will be calling `Commit()`
and then `Rollback()`, which causes an error. So you would need to keep track
if you have actually called `Commit()`

	func DatabaseOperation(db *sql.DB) {
		tx := db.Begin()
		commited := false
		defer func() {
			if commited {
				return
			}
			tx.Rollback()
		}

		... // database operation that may fail
		if err != nil {
			return
		}

		tx.Commit()
		commited = true
	}

This is doable, but you probably don't want to do that all over the place.

This is where this package comes in. The `Guard` interface
specifies `Fire()` and `Cancel()`, which makes the above construct
easier:

	func DatabaseOperation(db *sql.DB) {
		tx := db.Begin()
		g := guard.NewCB(func() error {
			return tx.Rollback()
		})
		defer g.Fire()

		... // database operation that may fail
		if err != nil {
			return
		}

		if err := tx.Commit(); err != nil {
			// If the commit is successful, we don't need to
			// rollback, so cancel the guard.
			return g.Cancel()
		}
	}

Once `Fire()` or `Cancel()` is called, the Guard never fires again, so
you can safely use it both in the success and failure cases.

Please also see: https://github.com/lestrrat/go-tx-guard

*/

import "errors"

// Guard is the interface for all guards. Most of your code should
// probably specify this as either return type or variable type.
type Guard interface {
	Fire() error
	Cancel() error
}

// ErrFired is returned if the guard has already been fired.
var ErrFired = errors.New("guard has already been fired")

// ErrCanceled is returned if the guard has already been canceled.
var ErrCanceled = errors.New("guard has already been canceled")

// Nil is a special guard that does nothing. Use it in tests or
// when you just need to pass a dummy guard to fulfill some function call.
var Nil nilGuard

type nilGuard struct{}

func (ng nilGuard) Fire() error   { return nil }
func (ng nilGuard) Cancel() error { return nil }

const (
	stFired    = 0x001
	stCanceled = 0x010
)

// CB is the most generic guard type, one that executes the given callback.
// While it's probably not a great thing to do, you can set the `OnFire`
// attribute at any given point to change the callback. You cannot
// recalibrate the guard so that it fires again.
type CB struct {
	state  int8
	OnFire func() error
}

// NewCB creates a new callback based guard.
func NewCB(onFire func() error) *CB {
	return &CB{
		state:  0,
		OnFire: onFire,
	}
}

// Fire executes the registered callback, only if the guard has not
// previously fired, and has not been canceled. The return value is
// nil if the callback successfully fired, and the callback did not
// return any errors.
func (c *CB) Fire() error {
	if c.state & stCanceled == stCanceled {
		return ErrCanceled
	}
	if c.state & stFired == stFired {
		return ErrFired
	}

	defer func() { c.state = c.state ^ stFired }()
	if cb := c.OnFire; cb != nil {
		return cb()
	}
	return nil
}

// Cancel sets the cancel flag so that subsequen calls to `Fire()`
// does not cause the callback to execute. It will return errors
// if the guard has already been fired or canceled.
func (c *CB) Cancel() error {
	if c.state & stCanceled == stCanceled {
		return ErrCanceled
	}
	if c.state & stFired == stFired {
		return ErrFired
	}

	c.state = c.state ^ stCanceled
	return nil
}