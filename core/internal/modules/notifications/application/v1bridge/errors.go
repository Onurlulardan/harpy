package v1bridge

import "errors"

var (
	ErrConfiguration = errors.New("notification V1 bridge configuration is invalid")
	ErrAuthority     = errors.New("notification V1 bridge authority is unavailable")
	ErrConflict      = errors.New("notification V1 bridge effect conflicts")
	ErrStorage       = errors.New("notification V1 bridge storage is unavailable")
)
