package menuhttp

import (
	"net/http"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	"github.com/google/uuid"
)

type Handler struct {
	runtime *cataloghttp.Runtime
	app     Application
}

type Route = cataloghttp.Route
type MenuOverrideInput = cataloghttp.MenuOverrideInput

var (
	ErrInvalid           = cataloghttp.ErrInvalid
	PermissionRead       = cataloghttp.PermissionRead
	PermissionMenuManage = cataloghttp.PermissionMenuManage
)

func New(runtime *cataloghttp.Runtime, app Application) *Handler {
	return &Handler{runtime: runtime, app: app}
}

func (handler *Handler) Ready() bool {
	return handler != nil && handler.runtime != nil && handler.runtime.Ready() &&
		cataloghttp.PortReady(handler.app)
}

func (handler *Handler) authorize(request *http.Request, permission cataloghttp.Permission,
	operation, purpose string,
) (cataloghttp.Access, error) {
	return handler.runtime.AccessFor(request, permission, operation, purpose)
}

func (handler *Handler) success(writer http.ResponseWriter, request *http.Request,
	status int, value any,
) {
	handler.runtime.Success(writer, request, status, value)
}

func (handler *Handler) failure(writer http.ResponseWriter, request *http.Request, err error) {
	handler.runtime.Failure(writer, request, err)
}

func route(method, pattern string, serve http.HandlerFunc) Route {
	return cataloghttp.NewRoute(method, pattern, serve)
}

func pathV7(request *http.Request, name string) (uuid.UUID, error) {
	return cataloghttp.PathV7(request, name)
}

func decodeBody[T any](request *http.Request, target *T) error {
	return cataloghttp.DecodeBody(request, target)
}

var requireNoQuery = cataloghttp.RequireNoQuery
var idempotencyKey = cataloghttp.IdempotencyKey
