package fieldregistry

type TypeID string

const (
	TypeText            TypeID = "text"
	TypeInteger         TypeID = "integer"
	TypeDecimal         TypeID = "decimal"
	TypeBoolean         TypeID = "boolean"
	TypeDate            TypeID = "date"
	TypeTimestamp       TypeID = "timestamp"
	TypeUUID            TypeID = "uuid"
	TypeEnum            TypeID = "enum"
	TypeRelation        TypeID = "relation"
	TypeTextList        TypeID = "text_list"
	TypeUUIDList        TypeID = "uuid_list"
	TypeJSON            TypeID = "json"
	TypeTime            TypeID = "time"
	TypeDocumentRef     TypeID = "document_ref"
	TypeDocumentRefList TypeID = "document_ref_list"
	TypeEnumList        TypeID = "enum_list"
)

type Role string

const (
	RoleStored      Role = "stored"
	RoleRelation    Role = "relation"
	RoleCalculation Role = "calculation"
)
