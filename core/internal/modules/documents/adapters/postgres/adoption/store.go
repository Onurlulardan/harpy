package adoption

import (
	"context"
	"fmt"
	"reflect"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type Row interface{ Scan(...any) error }
type QueryRow func(context.Context, string, ...any) Row
type Execute func(context.Context, string, ...any) error

type Store struct {
	tenantID domain.ID
	queryRow QueryRow
	execute  Execute
	allowed  func(context.Context) bool
}

func NewStore(tenantID domain.ID, queryRow QueryRow, execute Execute) (*Store, error) {
	if !tenantID.Valid() || queryRow == nil || execute == nil ||
		reflect.ValueOf(queryRow).IsNil() || reflect.ValueOf(execute).IsNil() {
		return nil, fmt.Errorf("documents adoption storage is invalid")
	}
	return &Store{tenantID: tenantID, queryRow: queryRow, execute: execute,
		allowed: foundation.DatabaseEffectsAllowed}, nil
}

func (storage *Store) usable(ctx context.Context) bool {
	return storage != nil && storage.tenantID.Valid() && storage.queryRow != nil &&
		storage.execute != nil && storage.allowed != nil && ctx != nil && ctx.Err() == nil &&
		storage.allowed(ctx)
}

func (storage *Store) owns(authority adoptionapp.Authority) bool {
	return storage != nil && authority.TenantID == storage.tenantID &&
		authority.RunID.UUIDv7() && authority.CutoverGeneration > 0 &&
		authority.ManifestSHA256 != [32]byte{}
}

var _ adoptionapp.Store = (*Store)(nil)
