package recordkernel

import (
	"unicode/utf8"
)

func validDecodedField(field Field, limits BudgetLimits) bool {
	isDecimal := field.TypeID == TypeDecimal
	if isDecimal != (field.Decimal != nil) || field.Decimal != nil &&
		(field.Decimal.Precision < 1 || field.Decimal.Precision > 1000 ||
			field.Decimal.Scale < 0 || field.Decimal.Scale > field.Decimal.Precision) {
		return false
	}
	isEnum := field.TypeID == TypeEnum || field.TypeID == TypeEnumList
	if isEnum != (field.EnumOptions != nil) || isEnum &&
		(len(field.EnumOptions) == 0 || uint64(len(field.EnumOptions)) >
			limits.EnumOptionsPerField) || !validEnumOptions(field.EnumOptions) {
		return false
	}
	isRelation := field.TypeID == TypeRelation &&
		field.Role == RoleRelation
	if isRelation != (field.Relation != nil) || field.Relation != nil &&
		(!stableUUIDv7(field.Relation.RelationID) ||
			!stableUUIDv7(field.Relation.TargetFormID) ||
			(field.Relation.Cardinality != "one" && field.Relation.Cardinality != "many") ||
			(field.Relation.DeletePolicy != "restrict" &&
				field.Relation.DeletePolicy != "cascade")) {
		return false
	}
	if field.TypeID != TypeText &&
		(field.TextMinimum != nil || field.TextMaximum != nil) {
		return false
	}
	if field.TextMinimum != nil || field.TextMaximum != nil {
		if field.TextMinimum == nil || field.TextMaximum == nil ||
			*field.TextMinimum < 0 || *field.TextMaximum < *field.TextMinimum ||
			*field.TextMaximum > MaximumFieldBytes {
			return false
		}
	}
	return true
}

func validEnumOptions(values []string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" || !utf8.ValidString(value) || len([]byte(value)) > 256 {
			return false
		}
		if _, duplicate := seen[value]; duplicate {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}
