package recoveryorchestration

import (
	"context"
	"fmt"
	environmentrecoveryv1 "github.com/Onurlulardan/harpy/contracts/environmentrecovery/v1"
	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
)

type discoveredRecoveryManifest struct {
	artifact environmentrecoveryv1.Artifact
	payload  recoveryset.ManifestPayload
}

func discoverManifest(
	ctx context.Context,
	storage RestoreStorage,
	environmentID, recoverySetID string,
	trust recoverysigningv1.Bundle,
) (discoveredRecoveryManifest, error) {
	documents, err := storage.ManifestVersions(
		ctx, environmentID, recoverySetID,
	)
	if err != nil {
		return discoveredRecoveryManifest{}, err
	}
	var found discoveredRecoveryManifest
	for _, document := range documents {
		artifact, payload, parseErr := parseRecoveryManifest(
			document, environmentID, recoverySetID, trust,
		)
		if parseErr != nil || found.artifact.SHA256() != "" &&
			found.artifact.SHA256() != artifact.SHA256() {
			return discoveredRecoveryManifest{},
				fmt.Errorf("recovery manifest versions conflict")
		}
		found = discoveredRecoveryManifest{
			artifact: artifact, payload: payload,
		}
	}
	if found.artifact.SHA256() == "" {
		return discoveredRecoveryManifest{},
			fmt.Errorf("recovery manifest is missing")
	}
	return found, nil
}
