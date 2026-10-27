package operator

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	platformauthority "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	installationmanifest "github.com/Onurlulardan/harpy/core/internal/platform/installationmanifest"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

const maximumAuthorityArtifactBytes = 64 * 1024

type operatorClock struct{}

func (operatorClock) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}

func BuildEnvironmentAuthority(
	ctx context.Context,
	settings config.OperatorInitialize,
	manifest installationmanifest.Manifest,
	pool *platformpostgres.Pool,
) (*EnvironmentAuthority, *platformauthority.GoogleClient, error) {
	if settings.DeploymentMode() != config.DeploymentSaaS {
		return nil, nil, nil
	}
	clock := operatorClock{}
	binding, err := platformauthority.ResolveSaaSBinding(
		ctx, pool, clock, manifest.EnvironmentIDString(),
	)
	if err != nil {
		return nil, nil, err
	}
	if settings.ProposerKMSKeyID().Value() != binding.SignerKID {
		return nil, nil, fmt.Errorf("configured SaaS signer differs from persisted binding")
	}
	identityPath := ""
	if identity, configured := settings.WorkloadIdentity(); configured {
		identityPath = identity.Path().Value()
	}
	client, err := platformauthority.OpenGoogleClient(ctx, identityPath)
	if err != nil {
		return nil, nil, err
	}
	planner, err := platformauthority.NewSaaSPlanner(ctx,
		platformauthority.SaaSPlannerDependencies{Pool: pool, Clock: clock,
			KMSKeyID: binding.SignerKID, KMSPublicKeySHA256: binding.SignerPublicKeySHA256,
			KMSClient: client})
	if err != nil {
		_ = client.Close()
		return nil, nil, err
	}
	ackTrust, err := readAuthorityArtifact(settings.AckTrustBundle().Path().Value())
	if err != nil {
		_ = client.Close()
		return nil, nil, err
	}
	releaseDocument, err := readAuthorityArtifact(settings.ReleaseTrustRoot().Path().Value())
	if err != nil {
		_ = client.Close()
		return nil, nil, err
	}
	defer clear(releaseDocument)
	releaseRoot, err := platformauthority.SelectSaaSReleaseRoot(
		releaseDocument, ackTrust, clock.Now(),
	)
	if err != nil {
		_ = client.Close()
		return nil, nil, err
	}
	acknowledger, err := platformauthority.NewSaaSAcknowledger(
		platformauthority.SaaSAcknowledgerDependencies{Pool: pool, Clock: clock,
			HTTPClient:  &http.Client{Timeout: 15 * time.Second},
			RegistryURL: settings.AuthorityRegistryURL().Value(), AckTrustBundle: ackTrust,
			ReleaseTrustRoot: releaseRoot})
	clear(ackTrust)
	if err != nil {
		_ = client.Close()
		return nil, nil, err
	}
	authority, err := NewEnvironmentAuthority(planner, acknowledger, binding)
	if err != nil {
		_ = client.Close()
		return nil, nil, err
	}
	return authority, client, nil
}

func readAuthorityArtifact(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open authority artifact: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() < 2 ||
		info.Size() > maximumAuthorityArtifactBytes {
		return nil, fmt.Errorf("authority artifact is invalid")
	}
	document := make([]byte, info.Size())
	if _, err := io.ReadFull(file, document); err != nil {
		clear(document)
		return nil, fmt.Errorf("read authority artifact: %w", err)
	}
	return document, nil
}
