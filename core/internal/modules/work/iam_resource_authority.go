package work

import (
	"context"

	workresourceapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/workresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type iamResourceAuthority struct{ iam workresourceapi.Authorizer }

func NewIAMResourceAuthority(
	iam workresourceapi.Authorizer,
) (ports.ResourceAuthority, error) {
	if iam == nil {
		return nil, ports.ErrUnavailable
	}
	return &iamResourceAuthority{iam: iam}, nil
}

func (authority *iamResourceAuthority) AuthorizeRead(
	ctx context.Context,
	local ports.Scope,
	request ports.ResourceAuthorityRequest,
) (bool, int64, error) {
	scope, ok := platformScope(local)
	if authority == nil || authority.iam == nil || !ok ||
		request.Owner != domain.SourceForms || !request.ID.Valid() ||
		!request.VersionID.Valid() {
		return false, 0, ports.ErrUnavailable
	}
	permission := workresourceapi.Request{Action: "read"}
	switch request.Kind {
	case "FORM":
		permission.Resource = "forms.definition"
	default:
		return false, 0, ports.ErrForbidden
	}
	decision, err := authority.iam.AuthorizeCurrent(ctx, scope, permission)
	if err != nil || !decision.Valid() {
		return false, 0, ports.ErrUnavailable
	}
	return decision.Allowed, decision.AuthorizationRevision, nil
}

var _ ports.ResourceAuthority = (*iamResourceAuthority)(nil)
