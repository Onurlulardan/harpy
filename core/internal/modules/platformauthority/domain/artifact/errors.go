package artifact

import "errors"

var (
	ErrInvalidDocument   = errors.New("protected authority document is invalid")
	ErrInvalidSignature  = errors.New("protected authority signature is invalid")
	ErrUntrustedKey      = errors.New("protected authority key is not trusted")
	ErrWrongRegistryKind = errors.New("protected authority registry kind is wrong")
	ErrWrongBinding      = errors.New("protected authority binding is wrong")
	ErrExpired           = errors.New("protected authority document expired")
	ErrGeneration        = errors.New("protected authority generation is invalid")
)
