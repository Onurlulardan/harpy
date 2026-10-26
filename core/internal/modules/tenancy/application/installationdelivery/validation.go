package installationdelivery

import (
	"context"
	"reflect"
)

func unusable(ctx context.Context, store Store) bool {
	if ctx == nil || ctx.Err() != nil || store == nil {
		return true
	}
	value := reflect.ValueOf(store)
	return value.Kind() == reflect.Pointer && value.IsNil()
}
