package licenseauthoritysource

import (
	"context"
	"time"

	licenseauthority "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
)

func ReadTrustBundle(
	ctx context.Context,
	path string,
	now time.Time,
) (licenseauthority.TrustBundle, error) {
	document, err := readStableArtifact(
		path, licenseauthority.MaximumDocumentBytes, false,
	)
	if err != nil {
		return licenseauthority.TrustBundle{}, err
	}
	return licenseauthority.ParseTrustBundle(ctx, document, now)
}

func ReadCredential(path string) ([]byte, error) {
	document, err := readStableArtifact(path, 256, true)
	if err != nil {
		return nil, err
	}
	if _, err = parseCredential(document); err != nil {
		return nil, err
	}
	return append([]byte(nil), document...), nil
}
