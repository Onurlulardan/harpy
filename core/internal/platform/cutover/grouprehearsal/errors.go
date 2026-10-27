package grouprehearsal

import "errors"

var (
	ErrInvalid = errors.New("cutover group rehearsal evidence is invalid")
	ErrFence   = errors.New("cutover group rehearsal requires a safety fence")
)
