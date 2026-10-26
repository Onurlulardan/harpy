package workresource

import (
	"context"
	"errors"
	"regexp"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

var (
	ErrInvalid     = errors.New("IAM Work resource permission request is invalid")
	ErrUnavailable = errors.New("IAM Work resource permission evidence is unavailable")
	keyPattern     = regexp.MustCompile(
		`^[a-z][a-z0-9]*([-_][a-z0-9]+)*(\.[a-z][a-z0-9]*([-_][a-z0-9]+)*)*$`,
	)
)

type Request struct {
	Resource string
	Action   string
}

func (request Request) Valid() bool {
	return len(request.Resource) <= 100 && len(request.Action) <= 100 &&
		keyPattern.MatchString(request.Resource) && keyPattern.MatchString(request.Action)
}

type Decision struct {
	Allowed               bool
	AuthorizationRevision int64
	AuthorizationClock    int64
}

func (decision Decision) Valid() bool {
	return decision.AuthorizationRevision > 0 && decision.AuthorizationClock > 0
}

type Authorizer interface {
	AuthorizeCurrent(
		context.Context, platformtx.TenantScope, Request,
	) (Decision, error)
}
