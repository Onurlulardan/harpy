package rowdecode

import "errors"

var (
	ErrInvalidFieldType = errors.New("REST Query row field type is invalid")
	ErrInvalidValue     = errors.New("REST Query PostgreSQL row value is invalid")
	ErrInvalidRow       = errors.New("REST Query PostgreSQL row is invalid")
	ErrDirectMarshal    = errors.New("REST Query row value cannot be marshaled directly")
)

const redacted = "<redacted>"
