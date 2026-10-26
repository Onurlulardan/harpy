package resolved

import (
	"strconv"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func specPathsCanonical(spec QuerySpec) bool {
	if spec.rootPath.value != "" {
		return false
	}
	if spec.record != nil {
		return recordPathsCanonical(*spec.record, spec.stableID)
	}
	return aggregatePathsCanonical(*spec.aggregate)
}

func projectionPathsCanonical(
	projections []Projection, stableID api.FieldDefinition, base string,
) bool {
	stable, err := stableID.ID()
	if err != nil {
		return false
	}
	indices := make(map[int]struct{}, len(projections))
	for _, projection := range projections {
		path := projection.path.value
		if path == base {
			fieldID, valid := fieldIdentifier(projection.field)
			if len(indices) != 0 && (!valid || fieldID != stable) {
				return false
			}
			continue
		}
		index, valid := structuralIndex(path, base, 50)
		if !valid {
			return false
		}
		if _, duplicate := indices[index]; duplicate {
			return false
		}
		indices[index] = struct{}{}
	}
	for index := 0; index < len(indices); index++ {
		if _, found := indices[index]; !found {
			return false
		}
	}
	if len(indices) == 0 {
		return true
	}
	for _, projection := range projections {
		if projection.path.value != base {
			continue
		}
		fieldID, valid := fieldIdentifier(projection.field)
		if !valid || fieldID != stable {
			return false
		}
	}
	return true
}

func sortPathsCanonical(sorts []Sort, base string) bool {
	indices := make(map[int]struct{}, len(sorts))
	for _, sort := range sorts {
		if sort.implicit {
			if sort.path.value != base {
				return false
			}
			continue
		}
		index, valid := structuralIndex(sort.path.value, base, 5)
		if !valid {
			return false
		}
		if _, duplicate := indices[index]; duplicate {
			return false
		}
		indices[index] = struct{}{}
	}
	for index := 0; index < len(indices); index++ {
		if _, found := indices[index]; !found {
			return false
		}
	}
	return true
}

func structuralIndex(path, base string, maximum int) (int, bool) {
	prefix := base + "/"
	if !strings.HasPrefix(path, prefix) {
		return 0, false
	}
	value := strings.TrimPrefix(path, prefix)
	if value == "" || len(value) > 1 && value[0] == '0' {
		return 0, false
	}
	index, err := strconv.Atoi(value)
	return index, err == nil && index >= 0 && index < maximum
}
