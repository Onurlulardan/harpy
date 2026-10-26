package aggregatepage

import "errors"

var (
	ErrContract      = errors.New("REST Query aggregate page contract is invalid")
	ErrInvalidPage   = errors.New("REST Query aggregate page is invalid")
	ErrDirectMarshal = errors.New("REST Query aggregate page cannot be marshaled directly")
)

const redacted = "<redacted>"
