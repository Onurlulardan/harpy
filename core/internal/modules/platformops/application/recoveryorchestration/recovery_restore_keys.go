package recoveryorchestration

import (
	"bytes"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/backupcrypto"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoverykeymaterial"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
)

type databaseRestoreKeys struct {
	material  *recoverykeymaterial.Material
	authority *databaseRestoreAuthority
}

func loadExistingRestoreKeys(
	path string,
	providerKind, providerHandle string,
	manifest recoveryset.ManifestPayload,
) (databaseRestoreKeys, error) {
	document, err := readRecoveryArtifact(
		path, recoverykeymaterial.MaximumDocumentBytes, true,
	)
	if err != nil {
		return databaseRestoreKeys{}, err
	}
	defer clear(document)
	material, err := recoverykeymaterial.Parse(
		document, recoverykeymaterial.Expected{
			Purpose:      recoverykeymaterial.PurposeComponentSealing,
			ProviderKind: providerKind, ProviderHandle: providerHandle,
			EnvironmentID: manifest.Identity.EnvironmentID,
			Generation:    manifest.RecoverySigningGeneration,
		},
	)
	if err != nil {
		return databaseRestoreKeys{},
			fmt.Errorf("restore DR keyring is invalid")
	}
	return databaseRestoreKeys{material: &material}, nil
}

func loadRehydrateRestoreKeys(
	authority *databaseRestoreAuthority,
) (databaseRestoreKeys, error) {
	if authority == nil || authority.kit == nil {
		return databaseRestoreKeys{},
			fmt.Errorf("rehydrate recovery keys are unavailable")
	}
	return databaseRestoreKeys{authority: authority}, nil
}

func (keys databaseRestoreKeys) decrypt(
	componentID string,
	environmentID, recoverySetID string,
	target *bytes.Buffer,
	source *bytes.Reader,
) (backupcrypto.Context, error) {
	if keys.material != nil {
		expected := backupcrypto.Context{
			EnvironmentID: environmentID, RecoverySetID: recoverySetID,
			ComponentID: componentID, DRKeyID: keys.material.DRKeyID(),
			RecoveryKeyID: keys.material.RecoveryKeyID(),
		}
		_, err := backupcrypto.DecryptExpected(
			source, target, expected, keys.material.DROpenKeys(),
		)
		return expected, err
	}
	_, binding, err := backupcrypto.DecryptBound(
		source, target, backupcrypto.BoundContext{
			EnvironmentID: environmentID, RecoverySetID: recoverySetID,
			ComponentID: componentID,
		}, keys.resolveRecoveryKey,
	)
	if err != nil {
		return backupcrypto.Context{}, err
	}
	return binding, nil
}

func (keys databaseRestoreKeys) resolveRecoveryKey(
	_ string,
	recoveryKeyID string,
) (backupcrypto.OpenKeys, bool) {
	if keys.authority == nil || keys.authority.kit == nil {
		return backupcrypto.OpenKeys{}, false
	}
	private, found := keys.authority.kit.RecoveryPrivateKey(recoveryKeyID)
	if !found {
		return backupcrypto.OpenKeys{}, false
	}
	return backupcrypto.OpenKeys{
		RecoveryPrivate: append([]byte(nil), private[:]...),
	}, true
}

func (keys *databaseRestoreKeys) clear() {
	if keys == nil || keys.material == nil {
		return
	}
	keys.material.Clear()
	keys.material = nil
}
