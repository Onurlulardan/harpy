package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/cutover/grouprehearsal"
)

func (recorder *groupRehearsalRecorder) Record(
	ctx context.Context,
	value grouprehearsal.PersistedRehearsal,
	digest [sha256.Size]byte,
) ([sha256.Size]byte, error) {
	members := value.Evidence.Members
	if recorder == nil || recorder.database == nil || len(members) != 2 {
		return [sha256.Size]byte{}, fmt.Errorf("billing Forms rehearsal recorder is invalid")
	}
	billing, forms := members[0], members[1]
	var persisted string
	err := recorder.database.QueryRow(ctx, recordBillingFormsRehearsalSQL,
		value.ReceiptID, value.Evidence.TenantID, value.Evidence.Generation,
		billing.OperationID, billing.RunID, billing.RollbackReceiptID,
		int64(billing.SourceCount), int64(billing.TargetCount),
		hex.EncodeToString(billing.AcceptanceSHA256[:]),
		hex.EncodeToString(billing.RollbackSHA256[:]),
		forms.OperationID, forms.RunID, forms.RollbackReceiptID,
		int64(forms.SourceCount), int64(forms.TargetCount),
		hex.EncodeToString(forms.AcceptanceSHA256[:]),
		hex.EncodeToString(forms.RollbackSHA256[:]),
		hex.EncodeToString(digest[:]), value.RecordedAt,
	).Scan(&persisted)
	if err != nil {
		return [sha256.Size]byte{}, fmt.Errorf("persist billing Forms rehearsal: %w", err)
	}
	decoded, decodeErr := hex.DecodeString(persisted)
	if decodeErr != nil || len(decoded) != sha256.Size {
		return [sha256.Size]byte{}, fmt.Errorf("persisted billing Forms digest is invalid")
	}
	var result [sha256.Size]byte
	copy(result[:], decoded)
	return result, nil
}
