package audit

import (
	auditpostgres "github.com/Onurlulardan/harpy/core/internal/modules/audit/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/audit/application"
)

type Application = application.Application
type AppendIntent = application.AppendIntent
type Permission = application.Permission
type IDSource = application.IDSource
type Clock = application.Clock

const PermissionRead = application.PermissionRead

var (
	ErrHTTPAuthentication = application.ErrAuthentication
	ErrHTTPForbidden      = application.ErrForbidden
	ErrHTTPUnavailable    = application.ErrUnavailable
)

func NewApplication(ids IDSource, clock Clock) (*Application, error) {
	return application.NewApplication(
		auditpostgres.NewBinder(), application.ContextAuthorizer{}, ids, clock)
}
