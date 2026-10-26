package application

import (
	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/google/uuid"
)

type SubmitCommand struct {
	ProposedJobID uuid.UUID
	Plan          schemav1.Plan
	PlanV2        *schemav2.Plan
}
