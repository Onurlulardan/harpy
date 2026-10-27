package schemav1

type Converter string

const (
	ConvertIdentity       Converter = "identity"
	ConvertIntegerDecimal Converter = "integer_to_decimal"
	ConvertIntegerText    Converter = "integer_to_text"
	ConvertDecimalText    Converter = "decimal_to_text"
	ConvertDateTimestamp  Converter = "date_to_timestamp"
	ConvertUUIDText       Converter = "uuid_to_text"
)

type VerifyCheck string

const (
	VerifyFormTable   VerifyCheck = "form_table"
	VerifyFields      VerifyCheck = "fields"
	VerifyRelations   VerifyCheck = "relations"
	VerifyIndexes     VerifyCheck = "indexes"
	VerifyConstraints VerifyCheck = "constraints"
)
