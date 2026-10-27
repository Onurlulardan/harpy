package archivedownload

import "errors"

var (
	ErrInvalid     = errors.New("documents archive download input is invalid")
	ErrForbidden   = errors.New("documents archive download is forbidden")
	ErrNotFound    = errors.New("documents archive artifact was not found")
	ErrExpired     = errors.New("documents archive artifact expired")
	ErrIntegrity   = errors.New("documents archive artifact integrity failed")
	ErrUnavailable = errors.New("documents archive download dependency is unavailable")
)
