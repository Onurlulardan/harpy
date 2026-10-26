package resource

import (
	"fmt"

	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	"github.com/google/uuid"
)

const storageOwner = "analytics"

type Verifier struct{}

func NewVerifier() *Verifier { return &Verifier{} }

func validUUIDv7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == 7 && value.Variant() == uuid.RFC4122
}

var _ publicationapp.ResourceVerifier = (*Verifier)(nil)
var _ dashboardapp.QueryVersionVerifier = (*Verifier)(nil)

var errResourceUnavailable = fmt.Errorf("analytics resource verifier is unavailable")
