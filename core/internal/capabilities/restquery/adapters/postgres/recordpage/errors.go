package recordpage

import "errors"

var (
	ErrContract      = errors.New("REST Query record page contract is invalid")
	ErrInvalidPage   = errors.New("REST Query record page is invalid")
	ErrDirectMarshal = errors.New("REST Query record page cannot be marshaled directly")
)

const redacted = "<redacted>"
