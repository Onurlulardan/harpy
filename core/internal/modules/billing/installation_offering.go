package billing

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var ErrInstallationOfferingUnavailable = errors.New(
	"billing installation offering is unavailable",
)

var installationOfferingSHA = regexp.MustCompile(`^[0-9a-f]{64}$`)

type InstallationOfferingReader interface {
	FetchInstallationOffering(context.Context, uuid.UUID, string) (InstallationOfferingPin, error)
}

type InstallationOfferingPin struct {
	VersionID                  uuid.UUID
	Revision, SourceSequence   int64
	SHA256, SourceSHA256, Mode string
	EffectiveAt                time.Time
	SourcePresent              bool
}

func (pin InstallationOfferingPin) IsValid() bool {
	if pin.VersionID == uuid.Nil || pin.Revision < 1 ||
		!installationOfferingSHA.MatchString(pin.SHA256) ||
		pin.EffectiveAt.IsZero() || pin.EffectiveAt.Location() != time.UTC {
		return false
	}
	if pin.Mode == "onpremise" {
		return pin.SourcePresent && pin.SourceSequence > 0 &&
			installationOfferingSHA.MatchString(pin.SourceSHA256)
	}
	return pin.Mode == "saas" && !pin.SourcePresent &&
		pin.SourceSequence == 0 && pin.SourceSHA256 == ""
}

func ReadInstallationOffering(
	ctx context.Context,
	reader InstallationOfferingReader,
	environmentID uuid.UUID,
	mode string,
) (InstallationOfferingPin, error) {
	if ctx == nil || ctx.Err() != nil || reader == nil ||
		environmentID == uuid.Nil || mode != "onpremise" && mode != "saas" {
		return InstallationOfferingPin{}, ErrInstallationOfferingUnavailable
	}
	pin, err := reader.FetchInstallationOffering(ctx, environmentID, mode)
	if err != nil {
		return InstallationOfferingPin{}, fmt.Errorf(
			"read installation offering: %w",
			errors.Join(ErrInstallationOfferingUnavailable, err),
		)
	}
	if !pin.IsValid() {
		return InstallationOfferingPin{}, ErrInstallationOfferingUnavailable
	}
	return pin, nil
}
