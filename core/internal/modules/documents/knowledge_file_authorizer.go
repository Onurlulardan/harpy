package documents

import (
	"context"

	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
	folderapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/folder"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
	knowledgeapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/knowledgefile"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type knowledgeFileEvidenceKey struct{}

type knowledgeFileAuthorizer struct{}

func (knowledgeFileAuthorizer) Authorize(ctx context.Context,
	scope platformtx.TenantScope, permission string,
) error {
	evidence, found := ctx.Value(knowledgeFileEvidenceKey{}).(knowledgeapp.Access)
	if !found || evidence.ActorID.String() != scope.ActorID().String() ||
		evidence.TenantID.String() != scope.TenantID().String() ||
		evidence.AuthorizationVersion < 1 || platformtx.NetworkAllowed(ctx) ||
		platformtx.RequireOwner(scope, "documents") != nil ||
		!knowledgeFilePermission(scope.OperationID(), permission) {
		return errApplicationForbidden
	}
	return nil
}

func knowledgeFilePermission(operation, permission string) bool {
	switch operation {
	case folderapp.OperationCreate:
		return permission == folderapp.PermissionCreate
	case uploadapp.OperationCreate, uploadapp.OperationPut:
		return permission == uploadapp.PermissionCreate || permission == uploadapp.PermissionVersion
	case downloadapp.OperationVersion:
		return permission == downloadapp.Permission
	case itemapp.OperationDelete:
		return permission == itemapp.PermissionDelete
	default:
		return false
	}
}

var _ foundation.Authorizer = knowledgeFileAuthorizer{}
