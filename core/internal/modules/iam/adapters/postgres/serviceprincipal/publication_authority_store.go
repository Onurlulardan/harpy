package serviceprincipal

import (
	"context"
	"encoding/hex"

	serviceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"
	"github.com/google/uuid"
)

func (store *publicationAuthorityStore) ManagePermissionID(
	ctx context.Context,
	scope tenantScope,
) (uuid.UUID, error) {
	bound, err := store.bindPublicationAuthority(ctx, scope)
	if err != nil {
		return uuid.Nil, err
	}
	var raw string
	if bound.database.QueryRow(ctx, publicationAuthorityPermissionSQL).Scan(&raw) != nil {
		return uuid.Nil, serviceapp.ErrEvidenceMissing
	}
	identifier, err := uuid.Parse(raw)
	if err != nil || identifier == uuid.Nil || identifier.String() != raw {
		return uuid.Nil, serviceapp.ErrEvidenceMissing
	}
	return identifier, nil
}

func (store *publicationAuthorityStore) PublicationPrincipal(
	ctx context.Context,
	scope tenantScope,
) (serviceapp.PublicationAuthorityPrincipal, error) {
	bound, err := store.bindPublicationAuthority(ctx, scope)
	if err != nil {
		return serviceapp.PublicationAuthorityPrincipal{}, err
	}
	var rawID, kind, state string
	if bound.database.QueryRow(ctx, publicationAuthorityPrincipalSQL).Scan(
		&rawID, &kind, &state,
	) != nil {
		return serviceapp.PublicationAuthorityPrincipal{}, serviceapp.ErrEvidenceMissing
	}
	if rawID == "" {
		return serviceapp.PublicationAuthorityPrincipal{}, nil
	}
	identifier, err := uuid.Parse(rawID)
	if err != nil || identifier == uuid.Nil || identifier.String() != rawID {
		return serviceapp.PublicationAuthorityPrincipal{}, serviceapp.ErrEvidenceMissing
	}
	return serviceapp.PublicationAuthorityPrincipal{
		ID: identifier, Found: true, Kind: kind, State: state,
	}, nil
}

func (store *publicationAuthorityStore) PublicationGrant(
	ctx context.Context,
	scope tenantScope,
	grantID uuid.UUID,
) (serviceapp.PublicationAuthorityGrant, error) {
	bound, err := store.bindPublicationAuthority(ctx, scope)
	if err != nil || grantID == uuid.Nil {
		return serviceapp.PublicationAuthorityGrant{}, serviceapp.ErrEvidenceMissing
	}
	return scanPublicationAuthorityGrant(
		bound.database.QueryRow(ctx, publicationAuthorityGrantSQL, grantID),
	)
}

func scanPublicationAuthorityGrant(
	row publicationAuthorityRow,
) (serviceapp.PublicationAuthorityGrant, error) {
	var rawID, rawPrincipal, capability, state, kind, owner, resourceKind string
	var rawResource, manifest string
	var revision int64
	if row == nil || row.Scan(&rawID, &rawPrincipal, &revision, &capability, &state,
		&kind, &owner, &resourceKind, &rawResource, &manifest) != nil {
		return serviceapp.PublicationAuthorityGrant{}, serviceapp.ErrEvidenceMissing
	}
	if rawID == "" {
		return serviceapp.PublicationAuthorityGrant{}, nil
	}
	id, idErr := uuid.Parse(rawID)
	principalID, principalErr := uuid.Parse(rawPrincipal)
	resourceID, resourceErr := uuid.Parse(rawResource)
	decoded, digestErr := hex.DecodeString(manifest)
	if idErr != nil || principalErr != nil || resourceErr != nil ||
		digestErr != nil || len(decoded) != 32 {
		return serviceapp.PublicationAuthorityGrant{}, serviceapp.ErrEvidenceMissing
	}
	var digest [32]byte
	copy(digest[:], decoded)
	return serviceapp.PublicationAuthorityGrant{
		PublicationAuthority: serviceapp.PublicationAuthority{
			ServicePrincipalID: principalID, GrantID: id, GrantRevision: revision,
		},
		Found: true, State: state, Capability: serviceapp.CapabilityKey(capability),
		Scope: serviceapp.ResourceScope{Kind: serviceapp.ScopeKind(kind), Owner: owner,
			ResourceKind: resourceKind, ResourceID: resourceID},
		ManifestSHA256: digest,
	}, nil
}
