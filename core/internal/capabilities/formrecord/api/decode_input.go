package recordkernel

import (
	"encoding/json"
	"fmt"
	"slices"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

func DecodeInput(
	catalog Catalog,
	action Action,
	authority MutationAuthority,
	document []byte,
) (DecodedInput, error) {
	if (action != ActionCreate && action != ActionUpdate) ||
		!authority.Valid() || len(document) < 2 ||
		uint64(len(document)) > catalog.limits.RecordInputBytes {
		return DecodedInput{}, ErrInputInvalid
	}
	if authority.Kind() == AuthorityWorkflow && !catalog.AllowsWorkflow(action) {
		return DecodedInput{}, ErrFieldForbidden
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		document, int(catalog.limits.RecordInputBytes),
	)
	if err != nil {
		return DecodedInput{}, ErrInputInvalid
	}
	var members map[string]json.RawMessage
	if json.Unmarshal(canonical, &members) != nil || members == nil {
		return DecodedInput{}, ErrInputInvalid
	}
	values := make(map[uuid.UUID]Value, len(members))
	relationTargets := 0
	names := make([]string, 0, len(members))
	for name := range members {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		raw := members[name]
		field, found := catalog.FieldByName(name)
		if !found {
			return DecodedInput{}, fmt.Errorf("%w: unknown field", ErrInputInvalid)
		}
		if !writeAllowed(field, action, authority.Kind()) {
			return DecodedInput{}, ErrFieldForbidden
		}
		value, targets, decodeErr := decodeInputValue(field, raw)
		if decodeErr != nil {
			return DecodedInput{}, decodeErr
		}
		relationTargets += targets
		if uint64(relationTargets) > catalog.limits.RelationTargetsPerMutation {
			return DecodedInput{}, ErrBudgetExceeded
		}
		values[field.ID] = value
	}
	fieldIDs := make([]uuid.UUID, 0, len(values))
	for fieldID := range values {
		fieldIDs = append(fieldIDs, fieldID)
	}
	slices.SortFunc(fieldIDs, compareUUID)
	requirement, err := NewRequirement(
		action, catalog.PublicationID(), catalog.FormID(), catalog.FormVersionID(), fieldIDs,
	)
	if err != nil {
		return DecodedInput{}, ErrCatalogInvalid
	}
	normalized, err := canonicalFieldIDDocument(values)
	if err != nil || uint64(len(normalized)) > catalog.limits.RecordInputBytes {
		return DecodedInput{}, ErrBudgetExceeded
	}
	return DecodedInput{action, values, fieldIDs, normalized, requirement}, nil
}

func decodeInputValue(field Field, raw []byte) (Value, int, error) {
	if field.TypeID == TypeRelation {
		return decodeRelation(field, raw)
	}
	value, err := DecodeValue(field, raw)
	return value, 0, err
}

func writeAllowed(
	field Field,
	action Action,
	authority AuthorityKind,
) bool {
	if field.Role == RoleCalculation || field.WritePolicy == WriteEngine ||
		action == ActionUpdate && field.Immutable {
		return false
	}
	if authority == AuthorityActor {
		return field.WritePolicy.allowsActor()
	}
	return authority == AuthorityWorkflow && field.WritePolicy.allowsWorkflow()
}
