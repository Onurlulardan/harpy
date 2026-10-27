package schemav2

type Converter string

const (
	ConvertIdentity       Converter = "identity_v1"
	ConvertIntegerDecimal Converter = "integer_to_decimal_scale0_v1"
	ConvertDecimalWiden   Converter = "decimal_widen_v1"
	ConvertIntegerText    Converter = "integer_to_text_v1"
	ConvertDecimalText    Converter = "decimal_to_text_v1"
	ConvertDateTimestamp  Converter = "date_to_timestamp_utc_v1"
	ConvertUUIDText       Converter = "uuid_to_text_v1"
)

type VerifyCheck string

const (
	VerifySystemColumns VerifyCheck = "system_columns"
	VerifyFormTable     VerifyCheck = "form_table"
	VerifyFields        VerifyCheck = "fields"
	VerifyRelations     VerifyCheck = "relations"
	VerifyIndexes       VerifyCheck = "indexes"
	VerifyEvolutionData VerifyCheck = "evolution_data"
	VerifyConstraints   VerifyCheck = "constraints"
)

type ChecksumAlgorithm string

const ChecksumXOR64PairV1 ChecksumAlgorithm = "xor64_pair_v1"
