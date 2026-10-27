package iam

import (
	"context"
	"fmt"
	"time"

	transferartifact "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/systemtransferartifact"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type systemTransferArtifactClock struct{}

func (systemTransferArtifactClock) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}

func NewSystemTransferArtifacts(
	ctx context.Context,
	settings config.IAMSystemTransferArtifacts,
	imports, exports SystemTransferArtifactObjectStore,
	spools *SystemTransferArtifactSpoolManager,
) (SystemTransferArtifacts, error) {
	if ctx == nil || ctx.Err() != nil || settings.Validate() != nil {
		return nil, ErrSystemTransferUnavailable
	}
	importSettings, err := settings.ImportSource()
	if err != nil {
		return nil, err
	}
	importKeys, err := newSystemTransferArtifactKeyring(importSettings)
	if err != nil {
		return nil, fmt.Errorf("open IAM import-source keyring: %w", err)
	}
	exportSettings, err := settings.Export()
	if err != nil {
		return nil, err
	}
	exportKeys, err := newSystemTransferArtifactKeyring(exportSettings)
	if err != nil {
		return nil, fmt.Errorf("open IAM export keyring: %w", err)
	}
	return transferartifact.NewRuntime(imports, exports, importKeys, exportKeys,
		spools, systemTransferArtifactClock{})
}
