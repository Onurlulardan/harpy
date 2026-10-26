package tabularmapping

import "errors"

var (
	ErrInvalid         = errors.New("tabular mapping is invalid")
	ErrRequiredMissing = errors.New("tabular mapping required source is missing")
)
