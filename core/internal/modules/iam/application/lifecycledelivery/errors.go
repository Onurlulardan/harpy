package lifecycledelivery

import "errors"

var (
	ErrInvalid             = errors.New("lifecycle delivery input is invalid")
	ErrArtifactUnavailable = errors.New("lifecycle delivery artifact is unavailable")
	ErrConflict            = errors.New("lifecycle delivery persisted evidence conflicts")
)
