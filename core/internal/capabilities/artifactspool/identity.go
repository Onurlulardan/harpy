package artifactspool

import (
	"fmt"

	"github.com/google/uuid"
)

type ReservationIdentity struct {
	ExportID          uuid.UUID `json:"export_id"`
	AttemptGeneration int64     `json:"attempt_generation"`
	LeaseOwnerID      uuid.UUID `json:"lease_owner_id"`
}

func (identity ReservationIdentity) validate() error {
	if identity.ExportID == uuid.Nil || identity.ExportID.Version() != 7 ||
		identity.AttemptGeneration < 1 || identity.AttemptGeneration > 3 ||
		identity.LeaseOwnerID == uuid.Nil {
		return fmt.Errorf("artifact spool reservation identity is invalid")
	}
	return nil
}
