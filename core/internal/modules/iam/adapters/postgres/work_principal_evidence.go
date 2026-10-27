package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

func scanWorkPrincipalEvidence(
	ctx context.Context,
	database workPrincipalQueryer,
	tenantID uuid.UUID,
	actorID uuid.UUID,
	targetKind string,
	targetID uuid.UUID,
) (WorkPrincipalEvidence, error) {
	var raw []byte
	var result WorkPrincipalEvidence
	err := database.QueryRow(
		ctx, workPrincipalEvidenceSQL, tenantID, actorID, targetKind, targetID,
	).Scan(
		&raw, &result.AuthorizationClock, &result.TargetExists, &result.TargetRevision,
	)
	if err != nil || result.AuthorizationClock < 0 ||
		json.Unmarshal(raw, &result.Principals) != nil {
		return WorkPrincipalEvidence{}, ErrWorkPrincipalRead
	}
	for _, principal := range result.Principals {
		if principal.ID == uuid.Nil || principal.Revision < 1 ||
			(principal.Kind != "USER" && principal.Kind != "ROLE" &&
				principal.Kind != "ORGANIZATION") {
			return WorkPrincipalEvidence{}, ErrWorkPrincipalRead
		}
	}
	return result, nil
}
