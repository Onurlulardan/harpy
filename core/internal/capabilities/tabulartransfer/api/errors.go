package api

import "errors"

var (
	ErrFormat      = errors.New("tabular transfer input format is invalid")
	ErrRowLimit    = errors.New("tabular transfer row limit is exceeded")
	ErrColumnLimit = errors.New("tabular transfer column limit is exceeded")
	ErrByteLimit   = errors.New("tabular transfer byte limit is exceeded")
)
