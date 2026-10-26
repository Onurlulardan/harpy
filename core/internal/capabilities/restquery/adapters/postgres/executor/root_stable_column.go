package executor

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func rootStableColumn(
	plan resolved.AuthorizedPlan,
	statement compiler.RecordStatement,
) (int, error) {
	spec, specErr := plan.Spec()
	stable, stableErr := spec.StableID()
	stableID, stableIDErr := stable.ID()
	columns, columnErr := statement.Columns()
	if specErr != nil || stableErr != nil || stableIDErr != nil || columnErr != nil {
		return 0, ErrContract
	}
	found := -1
	for index, column := range columns {
		identifier, err := column.FieldID()
		if err != nil {
			return 0, ErrContract
		}
		if identifier == stableID {
			if found >= 0 {
				return 0, ErrContract
			}
			found = index
		}
	}
	if found < 0 {
		return 0, ErrContract
	}
	return found, nil
}
