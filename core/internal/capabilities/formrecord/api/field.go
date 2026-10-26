package recordkernel

import (
	"github.com/google/uuid"
)

type DecimalProfile struct {
	Precision int
	Scale     int
}

type RelationProfile struct {
	RelationID   uuid.UUID `json:"relation_id"`
	TargetFormID uuid.UUID `json:"target_form_id"`
	Cardinality  string    `json:"cardinality"`
	DeletePolicy string    `json:"delete_policy"`
}

type Field struct {
	ID           uuid.UUID
	PublicName   string
	TypeID       TypeID
	Role         Role
	ValueProfile ValueProfile
	Nullable     bool
	WritePolicy  WritePolicy
	Generation   uint32
	Decimal      *DecimalProfile
	EnumOptions  []string
	Relation     *RelationProfile
	Required     bool
	Immutable    bool
	TextMinimum  *int
	TextMaximum  *int
}

func cloneField(source Field) Field {
	result := source
	result.EnumOptions = append([]string(nil), source.EnumOptions...)
	if source.Decimal != nil {
		profile := *source.Decimal
		result.Decimal = &profile
	}
	if source.Relation != nil {
		profile := *source.Relation
		result.Relation = &profile
	}
	if source.TextMinimum != nil {
		value := *source.TextMinimum
		result.TextMinimum = &value
	}
	if source.TextMaximum != nil {
		value := *source.TextMaximum
		result.TextMaximum = &value
	}
	return result
}
