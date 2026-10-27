package dependencyhttp

import (
	"context"
	"net/http"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	"github.com/google/uuid"
)

type Application interface {
	ListPackageVersionDependencies(context.Context, cataloghttp.Access, uuid.UUID, string) (
		[]cataloghttp.PackageDependency, error)
}

type Handler struct {
	runtime *cataloghttp.Runtime
	app     Application
}

type Route = cataloghttp.Route

func New(runtime *cataloghttp.Runtime, app Application) *Handler {
	return &Handler{runtime: runtime, app: app}
}

func (handler *Handler) Ready() bool {
	return handler != nil && handler.runtime != nil && handler.runtime.Ready() &&
		cataloghttp.PortReady(handler.app)
}

func (handler *Handler) Routes() []Route {
	if !handler.Ready() {
		return nil
	}
	return []Route{handler.ListRoute()}
}

func (handler *Handler) failure(writer http.ResponseWriter, request *http.Request, err error) {
	handler.runtime.Failure(writer, request, err)
}
