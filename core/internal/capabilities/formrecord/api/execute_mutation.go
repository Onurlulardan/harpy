package recordkernel

import (
	"github.com/google/uuid"
)

func ExecuteMutation(
	catalog Catalog,
	current Record,
	input DecodedInput,
) (EvaluatedRecord, error) {
	if catalog.sha256 == [32]byte{} || input.action != ActionCreate &&
		input.action != ActionUpdate {
		return EvaluatedRecord{}, ErrInputInvalid
	}
	if input.action == ActionCreate && len(current.values) != 0 {
		return EvaluatedRecord{}, ErrInputInvalid
	}
	if err := validateRecordShape(catalog, current); err != nil {
		return EvaluatedRecord{}, err
	}
	if input.action == ActionUpdate &&
		!recordStorageNormalized(catalog, current) {
		return EvaluatedRecord{}, ErrInputInvalid
	}
	record := current.clone()
	if record.values == nil {
		record.values = mapValues(input.values)
	} else {
		for fieldID, value := range input.values {
			record.values[fieldID] = value.clone()
		}
	}
	budget := newInstructionBudget(catalog)
	if input.action == ActionCreate {
		if err := applyDefaults(catalog, &record, &budget); err != nil {
			return EvaluatedRecord{}, err
		}
	}
	if err := applyCalculations(catalog, &record, &budget); err != nil {
		return EvaluatedRecord{}, err
	}
	materializeStorageShape(catalog, &record)
	if err := validateFinalRecord(catalog, record, &budget); err != nil {
		return EvaluatedRecord{}, err
	}
	canonical, err := canonicalFieldIDDocument(record.values)
	if err != nil || uint64(len(canonical)) > catalog.limits.RecordInputBytes ||
		len(canonical) > MaximumRecordBytes {
		return EvaluatedRecord{}, ErrBudgetExceeded
	}
	return EvaluatedRecord{record, canonical, budget.used}, nil
}

func mapValues(source map[uuid.UUID]Value) map[uuid.UUID]Value {
	result := make(map[uuid.UUID]Value, len(source))
	for fieldID, value := range source {
		result[fieldID] = value.clone()
	}
	return result
}
