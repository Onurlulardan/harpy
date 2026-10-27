package installationsecret

import "errors"

var (
	ErrInvalid     = errors.New("installation secret is invalid")
	ErrExists      = errors.New("installation secret already exists")
	ErrFenced      = errors.New("installation secret storage is fenced")
	ErrInterrupted = errors.New("installation secret storage was interrupted")
)
