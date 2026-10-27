package recoveryexecution

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

var componentKinds = map[string]bool{
	"database_archive":    true,
	"documents_inventory": true,
	"key_custody":         true,
	"owner_anchor":        true,
	"protected_authority": true,
}

func (service *Service) RecordComponent(
	ctx context.Context,
	record ports.RecoveryComponentRecord,
) error {
	if service == nil || ctx == nil || ctx.Err() != nil ||
		!validIdentity(ports.RecoveryExecutionIdentity{
			RecoverySetID: record.RecoverySetID,
			AttemptID:     record.AttemptID,
			LeaseToken:    record.LeaseToken,
		}) || !componentKinds[record.Kind] ||
		len(record.ObjectKey) < 1 || len(record.ObjectKey) > 1024 ||
		strings.ContainsAny(record.ObjectKey, "\x00\r\n") ||
		len(record.ObjectVersion) < 1 || len(record.ObjectVersion) > 1024 ||
		!validDigest(record.CipherSHA256) || record.CiphertextBytes < 1 {
		return ErrInvalid
	}
	changed, err := service.store.RecordRecoveryComponent(ctx, record)
	if err != nil {
		return ErrUnavailable
	}
	if !changed {
		return ErrConflict
	}
	return nil
}
