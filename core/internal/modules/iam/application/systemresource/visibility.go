package systemresource

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

type CatalogVisibility struct {
	read         bool
	confidential bool
	restricted   bool
	validated    bool
}

func NewCatalogVisibility(
	read bool,
	confidential bool,
	restricted bool,
) (CatalogVisibility, error) {
	visibility := CatalogVisibility{
		read: read, confidential: confidential,
		restricted: restricted, validated: true,
	}
	if visibility.Validate() != nil {
		return CatalogVisibility{}, publicresource.ErrContract
	}
	return visibility, nil
}

func (visibility CatalogVisibility) Validate() error {
	if !visibility.validated || !visibility.read {
		return publicresource.ErrContract
	}
	return nil
}

func (visibility CatalogVisibility) allows(classification profile.Classification) bool {
	if visibility.Validate() != nil {
		return false
	}
	switch classification {
	case profile.ClassificationPublic, profile.ClassificationInternal:
		return visibility.read
	case profile.ClassificationConfidential:
		return visibility.confidential
	case profile.ClassificationRestricted:
		return visibility.restricted
	default:
		return false
	}
}
