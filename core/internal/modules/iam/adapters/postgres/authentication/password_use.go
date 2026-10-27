package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func (store *store) UsePassword(
	ctx context.Context,
	scope authentication.CredentialScope,
	use authentication.PasswordUse,
) (authentication.PasswordUsed, error) {
	bound, err := store.bindLogin(ctx, scope)
	if err != nil {
		return authentication.PasswordUsed{}, err
	}
	return bound.usePassword(ctx, use)
}

func (bound *boundStore) usePassword(
	ctx context.Context,
	use authentication.PasswordUse,
) (authentication.PasswordUsed, error) {
	var salt, tag []byte
	var memory, iterations, parallelism any
	rehash := use.Rehash != nil
	if rehash {
		if !use.Rehash.Valid() {
			return authentication.PasswordUsed{}, errAuthenticationStorage
		}
		saltValue := use.Rehash.Salt()
		tagValue := use.Rehash.Tag()
		salt = append(salt, saltValue[:]...)
		tag = append(tag, tagValue[:]...)
		memory = int32(use.Rehash.MemoryKiB())
		iterations = int16(use.Rehash.Iterations())
		parallelism = int16(use.Rehash.Parallelism())
	}
	defer clear(salt)
	defer clear(tag)
	var result authentication.PasswordUsed
	err := bound.database.QueryRow(ctx, usePasswordSQL,
		use.UserID, use.ExpectedUserRevision, use.ExpectedTokenVersion,
		use.ExpectedAuthorizationVersion, use.ExpectedCredentialVersion,
		use.ExpectedCredentialRevision, rehash, nullableBytes(salt, rehash),
		nullableBytes(tag, rehash), memory, iterations, parallelism,
	).Scan(
		&result.UserRevision, &result.TokenVersion,
		&result.AuthorizationVersion, &result.CredentialVersion,
		&result.CredentialRevision,
	)
	if err != nil {
		return authentication.PasswordUsed{}, errAuthenticationStorage
	}
	return result, nil
}

func nullableBytes(value []byte, present bool) any {
	if !present {
		return nil
	}
	return value
}
