package guard

type causer interface {
	Cause() error
}

type guardFiredErrInterface interface {
	Fired() bool
}

type guardCanceledErrInterface interface {
	Canceled() bool
}

// IsFiredError returns the return value of `Fired()` method
// if the error implements it. Otherwise the method returns false
func IsFiredError(err error) bool {
	for err != nil {
		gerr, ok := err.(guardFiredErrInterface)
		if ok {
			return gerr.Fired()
		}

		cerr, ok := err.(causer)
		if ok {
			err = cerr.Cause()
		}
	}

	return false
}

// IsCanceledError returns the return value of `Canceled()` method
// if the error implements it. Otherwise the method returns false
func IsCanceledError(err error) bool {
	for err != nil {
		gerr, ok := err.(guardCanceledErrInterface)
		if ok {
			return gerr.Canceled()
		}

		cerr, ok := err.(causer)
		if ok {
			err = cerr.Cause()
		}
	}

	return false
}

