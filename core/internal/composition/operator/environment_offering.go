package operator

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/orchestration/environmentoperator"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type EnvironmentOffering struct {
	reader billing.InstallationOfferingReader
}

func NewEnvironmentOffering(
	pool *platformpostgres.Pool,
) (*EnvironmentOffering, error) {
	reader, err := billing.NewInstallationOfferingPlatformReader(pool)
	if err != nil {
		return nil, fmt.Errorf("bind installation offering reader: %w", err)
	}
	return &EnvironmentOffering{reader: reader}, nil
}

func (resolver *EnvironmentOffering) Resolve(
	ctx context.Context,
	request environmentoperator.OfferingRequest,
) (tenancymodule.BaseOfferingPin, error) {
	if resolver == nil || resolver.reader == nil || request.OperationID == [16]byte{} ||
		len(request.Manifest.CanonicalBytes()) == 0 {
		return tenancymodule.BaseOfferingPin{}, billing.ErrInstallationOfferingUnavailable
	}
	mode := string(request.Manifest.DeploymentMode())
	pin, err := billing.ReadInstallationOffering(ctx, resolver.reader,
		request.Manifest.EnvironmentID(), mode)
	if err != nil {
		return tenancymodule.BaseOfferingPin{}, err
	}
	result := tenancymodule.BaseOfferingPin{VersionID: tenancymodule.UUID(pin.VersionID),
		Revision: pin.Revision, SHA256: pin.SHA256, EffectiveAt: pin.EffectiveAt}
	if pin.SourcePresent {
		result.Source = tenancymodule.OfferingSourcePin{Present: true,
			Sequence: pin.SourceSequence, SHA256: pin.SourceSHA256}
	}
	return result, nil
}

var _ environmentoperator.Offering = (*EnvironmentOffering)(nil)
