package registryhttp

import (
	"context"
	"net/http"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	"github.com/google/uuid"
)

type Application interface {
	ListPackages(context.Context, cataloghttp.Access, int, *uuid.UUID) (
		[]cataloghttp.Package, *uuid.UUID, error)
	GetPackage(context.Context, cataloghttp.Access, uuid.UUID) (cataloghttp.Package, error)
	ListPackageVersions(context.Context, cataloghttp.Access, uuid.UUID) (
		[]cataloghttp.PackageVersion, error)
	GetPackageVersion(context.Context, cataloghttp.Access, uuid.UUID, string) (
		cataloghttp.PackageVersion, error)
}

type Handler struct {
	runtime *cataloghttp.Runtime
	app     Application
}

type Route = cataloghttp.Route
type Package = cataloghttp.Package
type PackageVersion = cataloghttp.PackageVersion

var PermissionRead = cataloghttp.PermissionRead

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

func pathPackageID(request *http.Request, name string) (uuid.UUID, error) {
	return cataloghttp.PathPackageID(request, name)
}

var requireNoQuery = cataloghttp.RequireNoQuery
var packagePageInput = cataloghttp.PackagePageInput
var pathVersion = cataloghttp.PathVersion

func (handler *Handler) Routes() []Route {
	if !handler.Ready() {
		return nil
	}
	return []Route{
		handler.ListPackagesRoute(), handler.GetPackageRoute(),
		handler.CompilePackageVersionsRoute(), handler.ResolvePackageVersionRoute(),
	}
}
