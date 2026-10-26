package recoveryanchorpostgres

import (
	"context"
	_ "embed"

	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
	"github.com/google/uuid"
)

//go:embed sql/platform/load_registration.sql
var loadRegistrationSQL string

func (storage *store) LoadRegistration(
	ctx context.Context,
	scope tenantScope,
	subject anchordomain.Subject,
	operationID string,
) (anchordomain.RegistrationLookup, error) {
	id, err := uuid.Parse(operationID)
	if err != nil {
		return anchordomain.RegistrationLookup{}, anchordomain.ErrInvalid
	}
	bound, err := storage.bind(ctx, scope, commandMode)
	if err != nil || scope.ActorID() != subject.ActorID || scope.TenantID() != subject.TenantID {
		return anchordomain.RegistrationLookup{}, anchordomain.ErrInvalid
	}
	var result anchordomain.RegistrationLookup
	var origins, digest []byte
	err = bound.database.QueryRow(ctx, loadRegistrationSQL,
		id, subject.TenantID, subject.ActorID).Scan(
		&result.OperationID, &result.AnchorID, &result.Challenge, &result.RPID,
		&origins, &digest, &result.ExpiresAt, &result.State)
	if err != nil {
		return anchordomain.RegistrationLookup{}, storeError(err, storage.rowNotFound)
	}
	result.Origins, err = decodeOrigins(origins)
	if err != nil || !copyDigest(&result.OriginsSHA256, digest) {
		return anchordomain.RegistrationLookup{}, anchordomain.ErrDependency
	}
	return result, nil
}
