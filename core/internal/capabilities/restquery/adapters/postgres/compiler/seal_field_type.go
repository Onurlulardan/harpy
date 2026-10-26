package compiler

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func writeFieldType(writer *sealWriter, fieldType api.FieldType) error {
	if writer == nil || fieldType.Validate() != nil {
		return ErrContract
	}
	kind, err := fieldType.Kind()
	if err != nil {
		return ErrContract
	}
	writer.writeInt(int64(kind))
	switch kind {
	case api.LogicalTypeText, api.LogicalTypeTextList:
		return writeTextProfile(writer, fieldType)
	case api.LogicalTypeInteger:
		profile, found, profileErr := fieldType.Integer()
		if profileErr != nil || !found {
			return ErrContract
		}
		minimum, minimumErr := profile.Minimum()
		maximum, maximumErr := profile.Maximum()
		if minimumErr != nil || maximumErr != nil {
			return ErrContract
		}
		writer.writeInt(minimum)
		writer.writeInt(maximum)
	case api.LogicalTypeDecimal:
		profile, found, profileErr := fieldType.Decimal()
		if profileErr != nil || !found {
			return ErrContract
		}
		precision, precisionErr := profile.Precision()
		scale, scaleErr := profile.Scale()
		if precisionErr != nil || scaleErr != nil {
			return ErrContract
		}
		writer.writeInt(int64(precision))
		writer.writeInt(int64(scale))
	case api.LogicalTypeEnum:
		return writeEnumProfile(writer, fieldType)
	}
	return nil
}

func writeTextProfile(writer *sealWriter, fieldType api.FieldType) error {
	profile, found, err := fieldType.Text()
	if err != nil || !found {
		return ErrContract
	}
	casefold, _, casefoldErr := profile.Casefold()
	accent, _, accentErr := profile.Accent()
	if casefoldErr != nil || accentErr != nil {
		return ErrContract
	}
	writer.writeInt(int64(casefold))
	writer.writeInt(int64(accent))
	return nil
}

func writeEnumProfile(writer *sealWriter, fieldType api.FieldType) error {
	profile, found, err := fieldType.Enum()
	if err != nil || !found {
		return ErrContract
	}
	version, versionErr := profile.Version()
	ordered, orderedErr := profile.Ordered()
	values, valuesErr := profile.Values()
	if versionErr != nil || orderedErr != nil || valuesErr != nil {
		return ErrContract
	}
	writer.writeBytes(version[:])
	writer.writeBool(ordered)
	writer.writeInt(int64(len(values)))
	for _, value := range values {
		code, codeErr := value.Code()
		ordinal, present, ordinalErr := value.Ordinal()
		if codeErr != nil || ordinalErr != nil || present != ordered {
			return ErrContract
		}
		writer.writeString(code)
		writer.writeInt(int64(ordinal))
	}
	return nil
}
