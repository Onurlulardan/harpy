package compiler

import (
	"bytes"
	"slices"

	"github.com/google/uuid"
)

const (
	maximumIncludeSourceIDs    = 1_000
	maximumIncludeCandidateIDs = 100_000
)

func normalizeIncludeUUIDs(
	values []uuid.UUID, maximum int,
) ([]uuid.UUID, error) {
	if values == nil || len(values) < 1 || len(values) > maximum || maximum < 1 {
		return nil, ErrContract
	}
	seen := make(map[uuid.UUID]struct{}, len(values))
	result := make([]uuid.UUID, 0, len(values))
	for _, identifier := range values {
		if identifier == uuid.Nil || identifier.Variant() != uuid.RFC4122 {
			return nil, ErrContract
		}
		if _, duplicate := seen[identifier]; duplicate {
			continue
		}
		seen[identifier] = struct{}{}
		result = append(result, identifier)
	}
	slices.SortFunc(result, func(left, right uuid.UUID) int {
		return bytes.Compare(left[:], right[:])
	})
	if len(result) < 1 || len(result) > maximum {
		return nil, ErrContract
	}
	return result, nil
}
