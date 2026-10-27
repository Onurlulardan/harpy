package systemresource

type RelationValue struct {
	field     string
	multiple  bool
	summaries []RelationSummary
	validated bool
}

func NewToOneRelation(
	field string,
	summary *RelationSummary,
) (RelationValue, error) {
	summaries := []RelationSummary{}
	if summary != nil {
		summaries = append(summaries, *summary)
	}
	return newRelationValue(field, false, summaries)
}

func NewToManyRelation(
	field string,
	summaries []RelationSummary,
) (RelationValue, error) {
	return newRelationValue(field, true, summaries)
}

func newRelationValue(
	field string,
	multiple bool,
	summaries []RelationSummary,
) (RelationValue, error) {
	value := RelationValue{
		field: field, multiple: multiple,
		summaries: append([]RelationSummary(nil), summaries...), validated: true,
	}
	if value.Validate() != nil {
		return RelationValue{}, ErrContract
	}
	return value, nil
}

func (value RelationValue) Validate() error {
	if !value.validated || value.field == "" ||
		!value.multiple && len(value.summaries) > 1 {
		return ErrContract
	}
	seen := make(map[string]struct{}, len(value.summaries))
	for _, summary := range value.summaries {
		id, _, _, err := summary.Values()
		if err != nil {
			return ErrContract
		}
		if _, duplicate := seen[id.String()]; duplicate {
			return ErrContract
		}
		seen[id.String()] = struct{}{}
	}
	return nil
}

func (value RelationValue) Values() (
	string,
	bool,
	[]RelationSummary,
	error,
) {
	if value.Validate() != nil {
		return "", false, nil, ErrContract
	}
	return value.field, value.multiple,
		append([]RelationSummary(nil), value.summaries...), nil
}

type RelationProjection struct {
	resource  Resource
	values    []RelationValue
	validated bool
}

func NewRelationProjection(
	resource Resource,
	values []RelationValue,
) (RelationProjection, error) {
	owned := make([]RelationValue, len(values))
	copy(owned, values)
	projection := RelationProjection{
		resource: resource,
		values:   owned, validated: true,
	}
	if projection.Validate() != nil {
		return RelationProjection{}, ErrContract
	}
	return projection, nil
}

func EmptyRelationProjection(resource Resource) (RelationProjection, error) {
	return NewRelationProjection(resource, []RelationValue{})
}

func (projection RelationProjection) Validate() error {
	if !projection.validated || !projection.resource.Valid() ||
		projection.values == nil {
		return ErrContract
	}
	seen := make(map[string]struct{}, len(projection.values))
	for _, value := range projection.values {
		field, _, _, err := value.Values()
		if err != nil || !relationFieldAllowed(projection.resource, field) {
			return ErrContract
		}
		if _, duplicate := seen[field]; duplicate {
			return ErrContract
		}
		seen[field] = struct{}{}
	}
	return nil
}

func (projection RelationProjection) Values() (
	Resource,
	[]RelationValue,
	error,
) {
	if projection.Validate() != nil {
		return "", nil, ErrContract
	}
	return projection.resource,
		append([]RelationValue(nil), projection.values...), nil
}

func relationFieldAllowed(resource Resource, field string) bool {
	switch resource {
	case Users:
		return field == "role_ids" || field == "organization_ids"
	case Roles:
		return field == "organization_id" || field == "user_ids"
	case Organizations:
		return field == "parent_id" || field == "user_ids" ||
			field == "role_ids" || field == "owner_ids"
	default:
		return false
	}
}
