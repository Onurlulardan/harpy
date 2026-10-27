package invitationdelivery

import "errors"

var (
	ErrInvalid             = errors.New("invitation delivery input is invalid")
	ErrArtifactUnavailable = errors.New("invitation delivery artifact is unavailable")
	ErrClaimHeld           = errors.New("invitation delivery artifact is already claimed")
	ErrConflict            = errors.New("invitation delivery state conflicts with persisted evidence")
)
