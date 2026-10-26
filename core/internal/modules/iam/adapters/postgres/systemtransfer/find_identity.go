package systemtransfer

import (
	"context"
	"encoding/json"
	"errors"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	identifierapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identifier"
	systemtransferapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

func (store *WorkerStore) FindIdentity(ctx context.Context,
	scope tenantScope, _ publicresource.ActorEvidence,
	manifest systemtransferapp.IdentityManifest, values publicresource.FieldValues,
) (systemtransferapp.IdentityMatch, error) {
	wanted, err := systemtransferapp.ManifestFor(manifest.Resource)
	if err != nil || wanted.Digest != manifest.Digest || values.Validate() != nil {
		return systemtransferapp.IdentityMatch{}, systemtransferapp.ErrInvalidAttempt
	}
	arguments, query, err := identityLookupArguments(scope, manifest.Resource, values)
	if err != nil {
		return systemtransferapp.IdentityMatch{}, err
	}
	bound, err := store.bind(ctx, scope)
	if err != nil {
		return systemtransferapp.IdentityMatch{}, err
	}
	var id uuid.UUID
	var revision int64
	err = bound.database.QueryRow(ctx, query, arguments...).Scan(&id, &revision)
	if bound.notFound != nil && bound.notFound(err) {
		return systemtransferapp.IdentityMatch{}, nil
	}
	if err != nil || id == uuid.Nil || revision < 1 {
		return systemtransferapp.IdentityMatch{}, publicresource.ErrUnavailable
	}
	return systemtransferapp.IdentityMatch{
		Found: true, ID: publicresource.ID(id), Revision: revision,
	}, nil
}

func identityLookupArguments(scope tenantScope,
	resource publicresource.Resource, values publicresource.FieldValues) ([]any, string, error) {
	items, err := values.Values()
	if err != nil {
		return nil, "", systemtransferapp.ErrInvalidAttempt
	}
	read := func(field string) (string, error) {
		id, parseErr := publicresource.ParseID(field)
		raw, found := items[id]
		var value string
		if parseErr != nil || !found || json.Unmarshal(raw, &value) != nil || value == "" {
			return "", systemtransferapp.ErrInvalidAttempt
		}
		return value, nil
	}
	switch resource {
	case publicresource.Users:
		email, readErr := read(publicresource.UserEmailFieldID)
		normalized, normalizeErr := identifierapp.NormalizeEmailV1(email)
		if readErr != nil || normalizeErr != nil {
			return nil, "", errors.Join(readErr, normalizeErr)
		}
		return []any{normalized.Normalized()}, findImportUserSQL, nil
	case publicresource.Roles:
		return roleLookupArguments(scope, items, read)
	case publicresource.Organizations:
		slug, readErr := read(publicresource.OrganizationSlugFieldID)
		return []any{slug}, findImportOrganizationSQL, readErr
	default:
		return nil, "", systemtransferapp.ErrInvalidAttempt
	}
}

func roleLookupArguments(scope tenantScope,
	items map[publicresource.ID]json.RawMessage,
	read func(string) (string, error)) ([]any, string, error) {
	kind, err := read(publicresource.RoleScopeKindFieldID)
	if err != nil || kind != "tenant" && kind != "organization" {
		return nil, "", systemtransferapp.ErrInvalidAttempt
	}
	name, err := read(publicresource.RoleNameFieldID)
	if err != nil {
		return nil, "", err
	}
	scopeID := scope.TenantID()
	if kind == "organization" {
		wire, readErr := read(publicresource.RoleScopeIDFieldID)
		parsed, parseErr := uuid.Parse(wire)
		if readErr != nil || parseErr != nil || parsed == uuid.Nil {
			return nil, "", systemtransferapp.ErrInvalidAttempt
		}
		scopeID = parsed
	}
	_ = items
	return []any{kind, scopeID, name}, findImportRoleSQL, nil
}

var _ systemtransferapp.IdentityLookup = (*WorkerStore)(nil)
