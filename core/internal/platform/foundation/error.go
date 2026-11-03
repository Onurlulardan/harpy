package foundation

import (
	"errors"
	"fmt"
)

type Error struct {
	Op        string
	Database  string
	Retryable bool
	Secrets   []string
	Err       error
}

func (e *Error) Error() string {
	cause := ""
	if e.Err != nil {
		cause = Redact(e.Err.Error(), e.Secrets)
	}
	return fmt.Sprintf(
		"foundation op=%s database=%s retryable=%t: %s",
		e.Op, e.Database, e.Retryable, cause,
	)
}

func (e *Error) Unwrap() error { return e.Err }

func failed(op string, cfg Config, retryable bool, err error) error {
	if err == nil {
		return nil
	}
	var already *Error
	if asError(err, &already) {
		return err
	}
	return &Error{
		Op: op, Database: cfg.Database, Retryable: retryable,
		Secrets: cfg.Secrets(), Err: err,
	}
}

func includeSecret(err error, secret string) error {
	var typed *Error
	if secret == "" || !errors.As(err, &typed) {
		return err
	}
	clone := *typed
	clone.Secrets = append(append([]string{}, typed.Secrets...), secret)
	return &clone
}

func asError(err error, target **Error) bool {
	return errors.As(err, target)
}
