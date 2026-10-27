//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"

	profilepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/identity/profile"
	userpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/identity/user"
	userapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/user"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	identityports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/identity"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type publicationRaceWriterResult struct {
	id       uuid.UUID
	revision int64
	replayed bool
	attempts int
}

type publicationRaceProfiles struct {
	identityports.Profile[platformtx.TenantScope]
	prepareErr error
}

func (profiles *publicationRaceProfiles) PreparePatch(
	ctx context.Context, scope platformtx.TenantScope,
	resource identitydomain.ProfileResource, id identitydomain.IdentityID, patch []byte,
) (identitydomain.ProfileDocument, error) {
	document, err := profiles.Profile.PreparePatch(ctx, scope, resource, id, patch)
	profiles.prepareErr = err
	return document, err
}

func executePublicationRaceWriter(
	ctx context.Context, runtime publicationRuntime, tenantID uuid.UUID,
	expectedRevision int64, patch []byte, operationID uuid.UUID,
) (publicationRaceWriterResult, error) {
	profiles, err := profilepostgres.NewStore()
	if err != nil {
		return publicationRaceWriterResult{}, err
	}
	users, err := userpostgres.NewStore(profiles)
	if err != nil {
		return publicationRaceWriterResult{}, err
	}
	trackedProfiles := &publicationRaceProfiles{Profile: profiles}
	commands, err := userapp.NewCommands(users, trackedProfiles)
	if err != nil {
		return publicationRaceWriterResult{}, err
	}
	generations, err := runtime.pool.ResolveSnapshotDomainGenerations(
		ctx, tenantID, []string{"iam", "system_resource_layout"},
	)
	if err != nil {
		return publicationRaceWriterResult{}, err
	}
	actorID := uuid.MustParse(publicationActorID)
	request, err := platformtx.AdmitCommandScope(
		actorID, tenantID, "iam.identity.user.update",
		"iam_system_resources", generations["iam"], generations,
	)
	if err != nil {
		return publicationRaceWriterResult{}, err
	}
	evidence := identitydomain.MutationEvidence{
		OperationID:   identitydomain.IdentityID(operationID),
		RequestSHA256: sha256.Sum256(patch),
		ActorID:       identitydomain.IdentityID(actorID),
		PermissionID: identitydomain.IdentityID(
			uuid.MustParse(publicationManagePerm),
		),
		ExpectedTokenVersion: 1, ExpectedAuthorizationVersion: 1,
	}
	var mutation identitydomain.MutationResult
	for attempt := 1; attempt <= 2; attempt++ {
		trackedProfiles.prepareErr = nil
		mutation = identitydomain.MutationResult{}
		err = runtime.coordinator.InTenant(
			ctx, request, func(scoped context.Context, scope platformtx.TenantScope) error {
				mutation, err = commands.UpdateUser(
					scoped, scope, userapp.UpdateUserCommand{
						ID:               identitydomain.IdentityID(actorID),
						ExpectedRevision: expectedRevision,
						ProfilePatch:     patch, Evidence: evidence,
					},
				)
				return err
			},
		)
		if err != nil && trackedProfiles.prepareErr != nil {
			err = fmt.Errorf(
				"prepare publication race profile: %w", trackedProfiles.prepareErr,
			)
		}
		result := publicationRaceWriterResult{
			id: uuid.UUID(mutation.ID), revision: mutation.Revision,
			replayed: mutation.Replayed, attempts: attempt,
		}
		if err == nil || !errors.Is(err, identitydomain.ErrStorage) || attempt == 2 {
			return result, err
		}
	}
	panic("unreachable publication race retry")
}
