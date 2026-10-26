package serviceprincipal

import (
	"context"

	serviceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type tenantScope = platformtx.TenantScope

type boundEvidenceReader struct {
	database platformpostgres.BoundExecutor
}

type boundCommandStore struct {
	database platformpostgres.BoundExecutor
}

type boundPublicationAuthorityStore struct {
	database platformpostgres.BoundExecutor
}

type evidenceReader struct {
	binder platformpostgres.OwnerBinder[*boundEvidenceReader]
	valid  bool
}

type commandStore struct {
	binder platformpostgres.OwnerBinder[*boundCommandStore]
	valid  bool
}

type publicationAuthorityStore struct {
	binder platformpostgres.OwnerBinder[*boundPublicationAuthorityStore]
	valid  bool
}

type commandRow = platformpostgres.Row
type publicationAuthorityRow = platformpostgres.Row

func NewEvidenceReader() (serviceapp.EvidencePort, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam",
		func(database platformpostgres.BoundExecutor) *boundEvidenceReader {
			return &boundEvidenceReader{database: database}
		},
	)
	if err != nil {
		return nil, errEvidenceRead
	}
	return &evidenceReader{binder: binder, valid: true}, nil
}

var _ serviceapp.EvidencePort = (*evidenceReader)(nil)

func NewCommandStore() (serviceapp.CommandPort, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam",
		func(database platformpostgres.BoundExecutor) *boundCommandStore {
			return &boundCommandStore{database: database}
		},
	)
	if err != nil {
		return nil, serviceapp.ErrEvidenceMissing
	}
	return &commandStore{binder: binder, valid: true}, nil
}

var _ serviceapp.CommandPort = (*commandStore)(nil)

func NewPublicationAuthorityStore() (serviceapp.PublicationAuthorityPort, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam",
		func(database platformpostgres.BoundExecutor) *boundPublicationAuthorityStore {
			return &boundPublicationAuthorityStore{database: database}
		},
	)
	if err != nil {
		return nil, serviceapp.ErrEvidenceMissing
	}
	return &publicationAuthorityStore{binder: binder, valid: true}, nil
}

var _ serviceapp.PublicationAuthorityPort = (*publicationAuthorityStore)(nil)

func (reader *evidenceReader) bind(scope tenantScope) (*boundEvidenceReader, error) {
	if reader == nil || !reader.valid {
		return nil, errEvidenceRead
	}
	bound, err := platformtx.Bind(scope, reader.binder)
	if err != nil || bound == nil {
		return nil, errEvidenceRead
	}
	return bound, nil
}

func (store *commandStore) bindCommand(scope tenantScope) (*boundCommandStore, error) {
	if store == nil || !store.valid {
		return nil, serviceapp.ErrEvidenceMissing
	}
	bound, err := platformtx.Bind(scope, store.binder)
	if err != nil || bound == nil {
		return nil, serviceapp.ErrEvidenceMissing
	}
	return bound, nil
}

func (store *publicationAuthorityStore) bindPublicationAuthority(
	ctx context.Context,
	scope tenantScope,
) (*boundPublicationAuthorityStore, error) {
	if store == nil || !store.valid || !evidenceScopeUsable(ctx, scope) {
		return nil, serviceapp.ErrEvidenceMissing
	}
	bound, err := platformtx.Bind(scope, store.binder)
	if err != nil || bound == nil {
		return nil, serviceapp.ErrEvidenceMissing
	}
	return bound, nil
}

func evidenceScopeUsable(ctx context.Context, scope tenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.Command) == nil
}
