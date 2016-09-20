package guard_test

import (
	"fmt"
	"sync"
	"testing"

	guard "github.com/lestrrat/go-simple-guard"
)

func fireGuard(g guard.Guard) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if errconv, ok := e.(error); ok {
				err = errconv
				return
			}

			err = fmt.Errorf("%s", e)
			return
		}
	}()
	defer g.Fire()
	return
}

func TestNilGuard(t *testing.T) {
	wg := &sync.WaitGroup{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := fireGuard(guard.Nil); err != nil {
				t.Errorf("Nil guard fire failed: %s", err)
			}
		}()
	}

	wg.Wait()
}

func TestDoubleFire(t *testing.T) {
	called := 0
	g := guard.Callback(func() error {
		called++
		return nil
	})
	defer func() {
		if called != 1 {
			t.Errorf("Expected guard to be called exactly once: %d", called)
		}
	}()

	for i := 0; i < 3; i++ {
		if i == 0 {
			if err := g.Fire(); err != nil {
				t.Errorf("first g.Fire() should not return an error on first invocation")
			}
		} else {
			err := g.Fire()
			if err == nil {
				t.Errorf("first g.Fire() should return an error on subsequent invocation")
				continue
			}

			if !guard.IsFiredError(err) {
				t.Errorf("first g.Fire() should return 'already fired' on subsequent invocation")
				t.Logf("Got %s", err)
			}
		}
	}
}

func TestDoubleCancel(t *testing.T) {
	called := 0
	g := guard.Callback(func() error {
		called++
		return nil
	})

	for i := 0; i < 3; i++ {
		if i == 0 {
			if err := g.Cancel(); err != nil {
				t.Errorf("first g.Cancel() should not return an error on first invocation")
			}
		} else {
			err := g.Cancel()
			if err == nil {
				t.Errorf("first g.Cancel() should return an error on subsequent invocation")
				continue
			}

			if !guard.IsCanceledError(err) {
				t.Errorf("first g.Cancel() should return ErrCanceled on subsequent invocation")
				t.Logf("Got %s", err)
			}
		}
	}
}
