package iam

import (
	"errors"
	"testing"

	serviceapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/serviceprincipal"
	serviceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"
)

func TestPublicationAuthorityErrorsCrossOnlyThePublicIAMContract(t *testing.T) {
	for _, internal := range []error{
		serviceapp.ErrGrantInactive,
		serviceapp.ErrGrantMismatch,
	} {
		if err := publicationAuthorityError(internal); !errors.Is(
			err, serviceapi.ErrPublicationAuthorityConflict,
		) {
			t.Fatalf("conflict mapping for %v = %v", internal, err)
		}
	}
	for _, internal := range []error{
		serviceapp.ErrInvalidGrant,
		serviceapp.ErrInvalidPrincipal,
		serviceapp.ErrEvidenceMissing,
		serviceapp.ErrChangeSink,
	} {
		if err := publicationAuthorityError(internal); !errors.Is(
			err, serviceapi.ErrPublicationAuthorityUnavailable,
		) {
			t.Fatalf("unavailable mapping for %v = %v", internal, err)
		}
	}
}
