package cmd

const (
	exitCodeNotFound = 1
	exitCodeUsage    = 2
	exitCodeFailure  = 3
)

type exitError struct {
	code int
	msg  string
}

func (e *exitError) Error() string {
	return e.msg
}

func (e *exitError) ExitCode() int {
	return e.code
}

func exitWith(message string, code int) error {
	return &exitError{code: code, msg: message}
}
