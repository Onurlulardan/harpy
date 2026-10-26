package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	"github.com/google/uuid"
)

func (authorizer *boundAuthorizer) readEvidence(
	ctx context.Context,
	tenantID uuid.UUID,
	actorID uuid.UUID,
	request authorization.Request,
) (authorization.Snapshot, error) {
	if authorizer == nil || authorizer.database == nil || ctx == nil ||
		ctx.Err() != nil || tenantID == uuid.Nil || actorID == uuid.Nil ||
		authorizer.query == "" || request.Validate() != nil {
		return authorization.Snapshot{}, errAuthorizationRead
	}
	var actorDocument, resourceDocument, actionDocument []byte
	var bindingsDocument, implicationsDocument []byte
	var clock int64
	arguments := []any{actorID, request.ResourceKey(), request.ActionKey()}
	if authorizer.includesTenant {
		arguments = []any{
			tenantID, actorID, request.ResourceKey(), request.ActionKey(),
		}
	}
	err := authorizer.database.QueryRow(ctx, authorizer.query, arguments...).Scan(
		&actorDocument, &resourceDocument, &actionDocument, &clock,
		&bindingsDocument, &implicationsDocument,
	)
	if err != nil {
		return authorization.Snapshot{}, fmt.Errorf("%w: %w", errAuthorizationRead, err)
	}
	return decodeAuthorizationSnapshot(
		actorDocument, resourceDocument, actionDocument, clock,
		bindingsDocument, implicationsDocument,
	)
}
