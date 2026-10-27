package catalogcomponentpostgres

import (
	"context"
	"fmt"
	"regexp"
)

var componentIDPattern = regexp.MustCompile(
	`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
)

func (storage *Store) AllowsMenuTarget(ctx context.Context, kind, target string) (bool, error) {
	if !storage.usable(ctx) || !componentIDPattern.MatchString(target) ||
		(kind != "FORM" && kind != "WORKFLOW" && kind != "DASHBOARD" && kind != "ENDPOINT") {
		return false, fmt.Errorf("catalog menu target is invalid")
	}
	var allowed bool
	if err := storage.queryRow(ctx, storage.queries.Allows, target, kind,
		storage.generation).Scan(&allowed); err != nil {
		return false, fmt.Errorf("catalog menu target is unavailable")
	}
	return allowed, nil
}
