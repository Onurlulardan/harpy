package recoveryorchestration

import (
	"bytes"
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
)

const maximumRestoreEvidenceBytes int64 = 2 * 1024 * 1024

func (executor *databaseRestoreExecutor) readRestoreComponent(
	ctx context.Context,
	kind string,
	identity recoveryset.ComponentIdentity,
	environmentID, recoverySetID string,
) ([]byte, error) {
	if identity.CiphertextBytes > maximumRestoreEvidenceBytes {
		return nil, fmt.Errorf("restore evidence component exceeds its limit")
	}
	var encrypted bytes.Buffer
	if err := executor.storage.RecoveryReader().DownloadVerified(
		ctx, restoreObjectIdentity(identity), &encrypted,
	); err != nil {
		return nil, err
	}
	var plaintext bytes.Buffer
	binding, err := executor.keys.decrypt(
		kind, environmentID, recoverySetID,
		&plaintext, bytes.NewReader(encrypted.Bytes()),
	)
	clear(encrypted.Bytes())
	if err != nil || binding.EnvironmentID != environmentID ||
		binding.RecoverySetID != recoverySetID ||
		binding.ComponentID != kind ||
		int64(plaintext.Len()) > maximumRestoreEvidenceBytes {
		clear(plaintext.Bytes())
		return nil, fmt.Errorf("restore evidence component binding differs")
	}
	return plaintext.Bytes(), nil
}
