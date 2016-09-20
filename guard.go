package guard

/*
Package guard provides a simple construct to help you write a RAII-like
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
		g := guard.Callback(func() error {
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

func (ng nilGuard) Fire() error   { return nil }
func (ng nilGuard) Cancel() error { return nil }

// Callback creates a new callback based guard.
func Callback(onFire func() error) *CB {
	return &CB{
		onFire: onFire,
	}
}

// NewCB is a deprecated constructor. Please use `Callback`
func NewCB(onFire func() error) *CB {
	return Callback(onFire)
}

func (c *CB) matchState(st int8) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.state&st == st
}

func (c *CB) setState(st int8) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.state = c.state ^ st
}

// Fire executes the registered callback, only if the guard has not
// previously fired, and has not been canceled. The return value is
// nil if the callback successfully fired, and the callback did not
// return any errors.
func (c *CB) Fire() error {
	if c.matchState(stCanceled) {
		return errCanceled
	}
	if c.matchState(stFired) {
		return errFired
	}

	defer c.setState(stFired)
	if cb := c.onFire; cb != nil {
		return cb()
	}
	return nil
}

// Cancel sets the cancel flag so that subsequen calls to `Fire()`
// does not cause the callback to execute. It will return errors
// if the guard has already been fired or canceled.
func (c *CB) Cancel() error {
	if c.matchState(stCanceled) {
		return errCanceled
	}
	if c.matchState(stFired) {
		return errFired
	}

	c.setState(stCanceled)
	return nil
}
