package compiler

import (
	"strconv"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
)

func validateTupleValues(
	columns []ResultColumn, values []cursor.LastValue,
) error {
	if validatePredecessorColumns(columns) != nil || len(values) != len(columns) {
		return ErrContract
	}
	for index, value := range values {
		if value.Validate() != nil {
			return ErrContract
		}
		null, err := value.IsNull()
		if err != nil || index == len(values)-1 && null {
			return ErrContract
		}
		if null {
			continue
		}
		resolvedValue, present, err := value.Value()
		kind, kindErr := resolvedValue.Kind()
		columnKind, columnErr := columns[index].LogicalType()
		_, bound, sorted, metadataErr := columns[index].SortMetadata()
		canonical, canonicalErr := tupleCanonicalBytes(resolvedValue, kind)
		if err != nil || !present || kindErr != nil || columnErr != nil ||
			metadataErr != nil || !sorted || canonicalErr != nil ||
			kind != columnKind || len(canonical) > bound {
			return ErrContract
		}
	}
	return nil
}

func tupleCanonicalBytes(value interface {
	Scalar() (string, bool, error)
	Boolean() (bool, bool, error)
}, kind api.LogicalType) ([]byte, error) {
	if kind == api.LogicalTypeBoolean {
		boolean, present, err := value.Boolean()
		if err != nil || !present {
			return nil, ErrContract
		}
		return []byte(strconv.FormatBool(boolean)), nil
	}
	scalar, present, err := value.Scalar()
	if err != nil || !present {
		return nil, ErrContract
	}
	return []byte(scalar), nil
}
