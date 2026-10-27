package postgres

import (
	"errors"

	"github.com/google/uuid"
)

var ErrWorkPrincipalRead = errors.New("IAM Work principal evidence read failed")

type WorkPrincipal struct {
	Kind     string    `json:"kind"`
	ID       uuid.UUID `json:"id"`
	Revision int64     `json:"revision"`
}

type WorkPrincipalEvidence struct {
	Principals         []WorkPrincipal
	AuthorizationClock int64
	TargetExists       bool
	TargetRevision     int64
}
