package presentation

import fieldregistry "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/fieldregistry"

type FieldMode string

const (
	FieldEditable FieldMode = "editable"
	FieldReadOnly FieldMode = "read_only"
)

func fieldProfileValid(
	typeID fieldregistry.TypeID,
	profile fieldregistry.ValueProfile,
) bool {
	wanted, valid := DefaultValueProfile(typeID)
	return valid && profile == wanted
}

func DefaultValueProfile(
	typeID fieldregistry.TypeID,
) (fieldregistry.ValueProfile, bool) {
	switch typeID {
	case fieldregistry.TypeText:
		return fieldregistry.ValueTextNFCV1, true
	case fieldregistry.TypeInteger:
		return fieldregistry.ValueIntegerBase10V1, true
	case fieldregistry.TypeDecimal:
		return fieldregistry.ValueDecimalBase10V1, true
	case fieldregistry.TypeBoolean:
		return fieldregistry.ValueBooleanV1, true
	case fieldregistry.TypeDate:
		return fieldregistry.ValueDateISO8601V1, true
	case fieldregistry.TypeTimestamp:
		return fieldregistry.ValueTimestampUTCMicrosecondV1, true
	case fieldregistry.TypeUUID:
		return fieldregistry.ValueUUIDRFC4122LowerV1, true
	case fieldregistry.TypeEnum:
		return fieldregistry.ValueEnumCodeV1, true
	case fieldregistry.TypeRelation:
		return fieldregistry.ValueRelationUUIDV1, true
	case fieldregistry.TypeTextList:
		return fieldregistry.ValueTextListNFCV1, true
	case fieldregistry.TypeUUIDList:
		return fieldregistry.ValueUUIDListV1, true
	case fieldregistry.TypeJSON:
		return fieldregistry.ValueJSONRFC8785V1, true
	case fieldregistry.TypeTime:
		return fieldregistry.ValueTimeMicrosecondV1, true
	case fieldregistry.TypeDocumentRef:
		return fieldregistry.ValueDocumentRefUUIDV1, true
	case fieldregistry.TypeDocumentRefList:
		return fieldregistry.ValueDocumentRefListV1, true
	case fieldregistry.TypeEnumList:
		return fieldregistry.ValueEnumListCodesV1, true
	default:
		return "", false
	}
}
