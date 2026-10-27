package worker

import (
	"context"
	"fmt"
	"path/filepath"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/artifactspool"
	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func openIAMSystemTransferArtifacts(ctx context.Context,
	settings config.IAMSystemTransferArtifacts) (iammodule.SystemTransferArtifacts, error) {
	imports, err := openIAMArtifactSide(ctx, settings.ImportSource,
		iammodule.SystemTransferImportObjectPrefix,
		iammodule.SystemTransferImportContract)
	if err != nil {
		return nil, fmt.Errorf("open IAM import artifacts: %w", err)
	}
	exports, err := openIAMArtifactSide(ctx, settings.Export,
		iammodule.SystemTransferExportObjectPrefix,
		iammodule.SystemTransferExportContract)
	if err != nil {
		return nil, fmt.Errorf("open IAM export artifacts: %w", err)
	}
	directory, err := settings.SpoolDirectory()
	if err != nil {
		return nil, err
	}
	spools, err := artifactspool.NewManager(filepath.Join(directory.Value(), "runtime"))
	if err != nil {
		return nil, fmt.Errorf("open IAM artifact spools: %w", err)
	}
	return iammodule.NewSystemTransferArtifacts(ctx, settings, imports, exports, spools)
}

type iamArtifactSettings func() (config.IAMTransferArtifactSettings, error)

func openIAMArtifactSide(ctx context.Context, load iamArtifactSettings,
	prefix, contract string) (iammodule.SystemTransferArtifactObjectStore, error) {
	settings, err := load()
	if err != nil {
		return nil, err
	}
	storage, err := settings.Storage()
	if err != nil {
		return nil, err
	}
	return artifactobject.OpenS3(ctx, storage, prefix, contract,
		iammodule.SystemTransferMaximumCiphertextBytes)
}
