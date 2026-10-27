package operator

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func issueRestoredLicenseSnapshot(
	ctx context.Context,
	settings config.OperatorLicenseReconcile,
	binding billing.LocalLicenseRestoreBinding,
	operationID billing.UUID,
	version string,
) ([]byte, error) {
	source, err := buildRestoredLicenseSource(settings.LicenseSource(), binding)
	if err != nil {
		return nil, err
	}
	var nonce [32]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return nil, fmt.Errorf("build restored license nonce: %w", err)
	}
	return billing.IssueLocalLicenseSnapshot(ctx, source, billing.LocalLicenseIssueRequest{
		OperationID: operationID, Binding: binding.Binding, Nonce: nonce,
		CurrentSequence: binding.Sequence, CurrentDigest: binding.SHA256,
		ComponentVersions: []billing.LocalLicenseComponentVersion{
			{Component: "core-operator", Version: version},
		},
	})
}

func buildRestoredLicenseSource(
	source config.LicenseSource,
	binding billing.LocalLicenseRestoreBinding,
) (billing.LocalLicenseSource, error) {
	switch binding.Binding.DeliveryKind {
	case billing.LocalLicenseDeliveryConnected:
		if source.Kind() != config.LicenseSourceConnected {
			return nil, fmt.Errorf("connected restored license source is unavailable")
		}
		credential, err := billing.ReadLocalLicenseCredential(
			source.CredentialFile().Path().Value(),
		)
		if err != nil {
			return nil, err
		}
		defer clear(credential)
		return billing.NewConnectedLicenseSource(
			source.IssuerURL().Value(), credential, http.DefaultTransport, 15*time.Second,
		)
	case billing.LocalLicenseDeliveryAirGap,
		billing.LocalLicenseDeliveryDevelopment:
		if source.Kind() != config.LicenseSourceSnapshot {
			return nil, fmt.Errorf("snapshot restored license source is unavailable")
		}
		return billing.NewLocalLicenseSnapshotFile(
			source.SnapshotFile().Path().Value(),
		)
	default:
		return nil, fmt.Errorf("restored license delivery is invalid")
	}
}
