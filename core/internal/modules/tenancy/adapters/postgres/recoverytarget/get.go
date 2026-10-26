package recoverytarget

import (
	"context"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/recoverytarget"
)

type Reader struct {
	within      recoveryTransaction
	rowNotFound func(error) bool
}

type recoveryRow interface{ Scan(...any) error }
type recoveryDatabase interface {
	QueryRow(context.Context, string, ...any) recoveryRow
}
type recoveryTransaction func(
	context.Context,
	func(recoveryDatabase) error,
) error

func (reader *Reader) Get(
	ctx context.Context,
	purpose string,
) (tenantapi.RecoveryTarget, bool, error) {
	var environmentID, tenantID string
	var slug, schemaName string
	var digest []byte
	found := true
	err := reader.within(ctx, func(database recoveryDatabase) error {
		err := database.QueryRow(ctx, getRecoveryTargetSQL, purpose).Scan(
			&environmentID, &tenantID, &slug, &schemaName, &digest,
		)
		if reader.rowNotFound(err) {
			found = false
			return nil
		}
		return err
	})
	if err != nil || !found || len(digest) != 32 {
		return tenantapi.RecoveryTarget{}, found, err
	}
	var sha [32]byte
	copy(sha[:], digest)
	target, err := tenantapi.NewRecoveryTarget(
		environmentID, tenantID, slug, schemaName, sha,
	)
	return target, err == nil, err
}

var _ recoverytarget.Reader = (*Reader)(nil)
