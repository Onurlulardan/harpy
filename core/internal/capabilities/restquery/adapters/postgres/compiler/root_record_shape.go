package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func recordShapeCount(query resolved.RecordQuery) (api.CountMode, error) {
	count, countErr := query.Count()
	if countErr != nil ||
		count != api.CountModeNone && count != api.CountModeExact {
		return 0, ErrContract
	}
	return count, nil
}

func validateRecordHasNoIncludes(query resolved.RecordQuery) error {
	includes, err := query.Includes()
	if err != nil || len(includes) != 0 {
		return ErrContract
	}
	return nil
}

func validateRecordPageShape(query resolved.RecordQuery) error {
	_, err := recordShapeCount(query)
	return err
}

func validateRecordCountShape(
	query resolved.RecordQuery, expected api.CountMode,
) error {
	count, err := recordShapeCount(query)
	if err != nil || count != expected {
		return ErrContract
	}
	return nil
}
