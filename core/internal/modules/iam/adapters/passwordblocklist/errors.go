package passwordblocklist

import "errors"

var (
	ErrUnavailable      = errors.New("password blocklist is unavailable")
	ErrCandidateInvalid = errors.New("password blocklist candidate is invalid")
)
