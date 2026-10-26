package recoveryanchorpostgres

import (
	"encoding/hex"
	"encoding/json"

	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
	"github.com/google/uuid"
)

func scanStaged(
	storage *store,
	row scanner,
	operationID anchordomain.ID,
	primaryAdmin [32]byte,
) (anchordomain.StagedMutation, error) {
	var result anchordomain.StagedMutation
	var anchorSet []byte
	var environment, ownerBinding, registryBinding uuid.UUID
	err := row.Scan(&result.AnchorID, &anchorSet, &result.Replayed,
		&environment, &result.Authority.DeploymentMode, &ownerBinding,
		&result.Authority.OwnerBindingGeneration, &registryBinding,
		&result.Authority.ProposalSignerGeneration, &result.Authority.ProposalSignerKID,
		&result.Authority.RecoveryTrustBundleSHA256)
	if err != nil {
		return anchordomain.StagedMutation{}, storeError(err, storage.rowNotFound)
	}
	if !copyDigest(&result.AnchorSetSHA256, anchorSet) ||
		result.Authority.RecoveryTrustBundleSHA256 == "" ||
		len(result.Authority.RecoveryTrustBundleSHA256) != 64 {
		return anchordomain.StagedMutation{}, anchordomain.ErrDependency
	}
	if _, err = hex.DecodeString(result.Authority.RecoveryTrustBundleSHA256); err != nil {
		return anchordomain.StagedMutation{}, anchordomain.ErrDependency
	}
	result.OperationID = operationID
	result.Authority.EnvironmentID = environment
	result.Authority.OwnerBindingID = ownerBinding
	result.Authority.RegistryBindingID = registryBinding
	result.Authority.PrimaryAdminBindingSetSHA256 = primaryAdmin
	result.Authority.RecoveryAnchorSetSHA256 = result.AnchorSetSHA256
	return result, nil
}

func decodeOrigins(raw []byte) ([]string, error) {
	var origins []string
	if json.Unmarshal(raw, &origins) != nil || len(origins) == 0 {
		return nil, anchordomain.ErrDependency
	}
	return origins, nil
}

func copyDigest(target *[32]byte, source []byte) bool {
	return len(source) == len(target) && copy(target[:], source) == len(target)
}
