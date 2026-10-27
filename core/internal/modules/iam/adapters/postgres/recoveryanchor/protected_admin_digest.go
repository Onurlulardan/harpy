package recoveryanchorpostgres

import (
	"context"
	_ "embed"

	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
)

//go:embed sql/tenant/protected_admin_digest.sql
var protectedAdminDigestSQL string

func (storage *store) protectedAdminDigest(
	ctx context.Context,
	bound *boundStore,
	environmentID anchordomain.ID,
	tenantID anchordomain.ID,
) ([32]byte, error) {
	var raw []byte
	if err := bound.database.QueryRow(
		ctx,
		protectedAdminDigestSQL,
		environmentID,
		tenantID,
	).Scan(&raw); err != nil {
		return [32]byte{}, storeError(err, storage.rowNotFound)
	}
	var result [32]byte
	if !copyDigest(&result, raw) {
		return [32]byte{}, anchordomain.ErrDependency
	}
	return result, nil
}
