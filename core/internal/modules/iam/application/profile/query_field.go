package profile

import (
	"github.com/google/uuid"

	restquery "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type QueryField struct {
	id           uuid.UUID
	publicName   string
	logicalType  restquery.LogicalType
	fieldType    restquery.FieldType
	class        Classification
	required     bool
	maximumBytes int
	operators    []restquery.Operator
	validated    bool
}

func newQueryField(field Field) (QueryField, error) {
	fieldType, err := queryFieldType(field)
	if err != nil || field.state != FieldActive {
		return QueryField{}, ErrQueryCapabilityRejected
	}
	logicalType, err := fieldType.Kind()
	if err != nil {
		return QueryField{}, ErrQueryCapabilityRejected
	}
	return QueryField{
		id: field.id, publicName: field.key, logicalType: logicalType,
		fieldType: fieldType, class: field.classification, required: field.required,
		maximumBytes: field.validation.maximumBytes,
		operators:    queryOperators(field), validated: true,
	}, nil
}

func (field QueryField) Validate() error {
	if !field.validated || field.id == uuid.Nil || field.publicName == "" ||
		!field.logicalType.Valid() || field.fieldType.Validate() != nil ||
		!field.class.Valid() || field.maximumBytes < 1 ||
		field.maximumBytes > MaximumDocumentBytes || len(field.operators) == 0 {
		return ErrQueryCapabilityRejected
	}
	return nil
}

func (field QueryField) FieldType() (restquery.FieldType, error) {
	if err := field.Validate(); err != nil {
		return restquery.FieldType{}, err
	}
	return field.fieldType, nil
}

func (field QueryField) Classification() (Classification, error) {
	if err := field.Validate(); err != nil {
		return "", err
	}
	return field.class, nil
}

func (field QueryField) Required() (bool, error) {
	if err := field.Validate(); err != nil {
		return false, err
	}
	return field.required, nil
}

func (field QueryField) MaximumBytes() (int, error) {
	if err := field.Validate(); err != nil {
		return 0, err
	}
	return field.maximumBytes, nil
}

func (field QueryField) ID() (uuid.UUID, error) {
	if err := field.Validate(); err != nil {
		return uuid.Nil, err
	}
	return field.id, nil
}
func (field QueryField) PublicName() (string, error) {
	if err := field.Validate(); err != nil {
		return "", err
	}
	return field.publicName, nil
}
func (field QueryField) LogicalType() (restquery.LogicalType, error) {
	if err := field.Validate(); err != nil {
		return 0, err
	}
	return field.logicalType, nil
}
func (field QueryField) Selectable() (bool, error) {
	if err := field.Validate(); err != nil {
		return false, err
	}
	return true, nil
}
func (field QueryField) Sortable() (bool, error) {
	if err := field.Validate(); err != nil {
		return false, err
	}
	return false, nil
}
func (field QueryField) Searchable() (bool, error) {
	if err := field.Validate(); err != nil {
		return false, err
	}
	return false, nil
}

func (field QueryField) Operators() ([]restquery.Operator, error) {
	if err := field.Validate(); err != nil {
		return nil, err
	}
	return append([]restquery.Operator(nil), field.operators...), nil
}

func (field QueryField) RequireOperator(operator restquery.Operator) error {
	if field.Validate() != nil {
		return ErrQueryCapabilityRejected
	}
	if _, valid := operator.Arity(); !valid {
		return ErrQueryCapabilityRejected
	}
	for _, allowed := range field.operators {
		if allowed == operator {
			return nil
		}
	}
	return ErrQueryCapabilityRejected
}

func (field QueryField) RequireSort() error {
	if field.Validate() != nil {
		return ErrQueryCapabilityRejected
	}
	return ErrQueryCapabilityRejected
}
func (field QueryField) RequireSearch() error {
	if field.Validate() != nil {
		return ErrQueryCapabilityRejected
	}
	return ErrQueryCapabilityRejected
}
