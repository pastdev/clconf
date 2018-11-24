package cmd

type exitError struct {
	wrapped  error
	exitCode int
	message  string
}

// NewExitErrorWrapper creates a new error that holds an exit code.
func NewExitErrorWrapper(exitCode int, wrapped error) error {
	return &exitError{
		exitCode: exitCode,
		wrapped:  wrapped,
	}
}

// NewExitError creates a new error that holds an exit code.
func NewExitError(exitCode int, message string) error {
	return &exitError{
		exitCode: exitCode,
		message:  message,
	}
}

func (e *exitError) Error() string {
	if e.wrapped == nil {
		return e.message
	}
	return e.wrapped.Error()
}
