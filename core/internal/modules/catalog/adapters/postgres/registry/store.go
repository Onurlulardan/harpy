package registry

import (
	"context"
	"regexp"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

const storageOwner = "catalog"

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
var slugPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,62}$`)
var storedVersionPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

type row interface{ Scan(...any) error }
type rows interface {
	Close()
	Err() error
	Next() bool
	Scan(...any) error
}

type store struct {
	queryRow func(context.Context, string, ...any) row
	query    func(context.Context, string, ...any) (rows, error)
	notFound func(error) bool
	allowed  func(context.Context) bool
}

func (storage *store) usable(ctx context.Context) bool {
	return storage != nil && storage.queryRow != nil && storage.query != nil &&
		storage.notFound != nil && storage.allowed != nil && ctx != nil &&
		ctx.Err() == nil && storage.allowed(ctx)
}

func validPackageID(raw string) bool {
	value, err := packagecontract.ParsePackageID(raw)
	return err == nil && value.String() == raw
}

func validStableID(raw string) bool {
	value, err := packagecontract.ParseStableID(raw)
	return err == nil && value.String() == raw
}
