package knowledgefile

import (
	"fmt"
	"reflect"

	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
	folderapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/folder"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
)

type Application struct {
	scopes    Scopes
	folders   *folderapp.Application
	uploads   *uploadapp.Application
	downloads *downloadapp.Application
	items     *itemapp.Application
	leases    LeaseBinder
	links     LinkBinder
	ids       foundation.IDSource
	clock     foundation.Clock
}

func New(scopes Scopes, folders *folderapp.Application,
	uploads *uploadapp.Application, downloads *downloadapp.Application,
	items *itemapp.Application,
	leases LeaseBinder, links LinkBinder, ids foundation.IDSource,
	clock foundation.Clock,
) (*Application, error) {
	dependencies := []any{scopes, folders, uploads, downloads, items, leases, links, ids, clock}
	for _, dependency := range dependencies {
		if missing(dependency) {
			return nil, fmt.Errorf("Documents knowledge-file dependencies are required")
		}
	}
	return &Application{scopes: scopes, folders: folders, uploads: uploads,
		downloads: downloads, items: items, leases: leases, links: links,
		ids: ids, clock: clock}, nil
}

func missing(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	switch seen.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return seen.IsNil()
	default:
		return false
	}
}
