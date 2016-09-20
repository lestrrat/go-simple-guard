# go-simple-guard

Simple guard construct

[![Build Status](https://travis-ci.org/lestrrat/go-simple-guard.png?branch=master)](https://travis-ci.org/lestrrat/go-simple-guard)

[![GoDoc](https://godoc.org/github.com/lestrrat/go-simple-guard?status.svg)](https://godoc.org/github.com/lestrrat/go-simple-guard)

# SYNOPSIS

```go
func Example() {
  g := guard.Callback(func() {
    // some piece of cleanup code
  })

  // Automatically call the callback upon
  // finishing the code
  defer g.Fire()

  if ... {
    // Maybe you don't want to fire the callback
    // anymore. We can cancel it too
    g.Cancel()
  }
}
```

# DESCRIPTION

Go has a `defer` builtin, but sometimes you need the ability to cancel it. The `Guard` construct wraps this, and allows you to safely call and cancel the callbacks.
