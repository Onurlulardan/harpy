package recoveryevidence

import (
	"context"
	"fmt"

	evidenceports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/recoveryevidence"
)

func (store *Store) ReadOwnerAnchors(
	ctx context.Context,
	environmentID string,
) ([]evidenceports.Anchor, error) {
	result := make([]evidenceports.Anchor, 0)
	err := store.within(ctx, func(database database) error {
		rows, err := database.Query(ctx, readOwnerAnchorsSQL, environmentID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var item evidenceports.Anchor
			if err = rows.Scan(
				&item.EnvironmentID, &item.TenantID, &item.OwnerUserID,
				&item.OwnerBindingID, &item.OwnerBindingGeneration,
				&item.AnchorID, &item.AnchorGeneration, &item.CredentialID,
				&item.PublicKeyCOSE, &item.AAGUID, &item.AttestationFormat,
				&item.AttestationType, &item.RPID, &item.AllowedOrigins,
				&item.AllowedOriginsSHA256, &item.UserVerification,
				&item.BackupEligible, &item.BackupState, &item.SignCount,
				&item.CounterPolicy, &item.ProtectedHeadGeneration,
				&item.ProtectedStateSHA256, &item.ProtectedAckSHA256,
			); err != nil {
				return err
			}
			result = append(result, item)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("read IAM recovery owner evidence: %w", err)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("IAM recovery owner evidence is unavailable")
	}
	return result, nil
}

var _ evidenceports.Reader = (*Store)(nil)
