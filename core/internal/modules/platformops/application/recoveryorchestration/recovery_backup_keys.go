package recoveryorchestration

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoverykeymaterial"
)

func loadDatabaseBackupKeyMaterial(
	providerKind, providerHandle string,
	environmentID string,
	generation int64,
) (recoverykeymaterial.Material, error) {
	path, err := recoveryProviderPath(providerKind, providerHandle)
	if err != nil {
		return recoverykeymaterial.Material{}, err
	}
	document, err := readRecoveryFile(
		path, recoverykeymaterial.MaximumDocumentBytes, true,
	)
	if err != nil {
		return recoverykeymaterial.Material{}, err
	}
	defer clear(document)
	material, err := recoverykeymaterial.Parse(
		document,
		recoverykeymaterial.Expected{
			Purpose:        recoverykeymaterial.PurposeComponentSealing,
			ProviderKind:   providerKind,
			ProviderHandle: providerHandle,
			EnvironmentID:  environmentID,
			Generation:     generation,
		},
	)
	if err != nil {
		return recoverykeymaterial.Material{},
			fmt.Errorf("recovery key provider material is invalid")
	}
	return material, nil
}
