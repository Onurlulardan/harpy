package recoveryorchestration

import "fmt"

func (executor *databaseRestoreExecutor) validateKeyCustodyEvidence(
	value restoreKeyCustodyEvidence,
	environmentID, recoverySetID string,
) error {
	if value.Schema != "recovery-key-custody-evidence.v1" ||
		value.EnvironmentID != environmentID ||
		value.RecoverySetID != recoverySetID ||
		value.ProviderKind == "" || value.ProviderHandle == "" ||
		value.ProviderGeneration < 1 ||
		value.ProviderMaterialSHA256 == "" ||
		value.DRKEKID == "" || value.RecoveryKeyID == "" {
		return fmt.Errorf("recovery key custody evidence differs")
	}
	if executor.keys.material != nil {
		material := executor.keys.material
		if value.ProviderKind != material.ProviderKind() ||
			value.ProviderHandle != material.ProviderHandle() ||
			value.ProviderGeneration != material.Generation() ||
			value.ProviderMaterialSHA256 != material.DocumentSHA256() ||
			value.DRKEKID != material.DRKeyID() ||
			value.RecoveryKeyID != material.RecoveryKeyID() {
			return fmt.Errorf("recovery key custody material differs")
		}
		return nil
	}
	if executor.authority == nil || executor.authority.kit == nil {
		return fmt.Errorf("recovery kit custody is unavailable")
	}
	private, found := executor.authority.kit.RecoveryPrivateKey(
		value.RecoveryKeyID,
	)
	if !found {
		return fmt.Errorf("recovery kit custody differs")
	}
	clear(private[:])
	return nil
}
