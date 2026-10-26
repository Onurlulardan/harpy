package api

func fieldCapabilityTypeShape(
	fieldType FieldType,
) (LogicalType, bool, bool, error) {
	if err := fieldType.Validate(); err != nil {
		return 0, false, false, err
	}
	kind, err := fieldType.Kind()
	if err != nil {
		return 0, false, false, err
	}
	orderedEnum := false
	if kind == LogicalTypeEnum {
		profile, found, profileErr := fieldType.Enum()
		if profileErr != nil || !found {
			return 0, false, false, errInvalidFieldCapabilities
		}
		orderedEnum, err = profile.Ordered()
		if err != nil {
			return 0, false, false, err
		}
	}
	casefoldText := false
	if kind == LogicalTypeText {
		profile, found, profileErr := fieldType.Text()
		if profileErr != nil || !found {
			return 0, false, false, errInvalidFieldCapabilities
		}
		_, casefoldText, err = profile.Casefold()
		if err != nil {
			return 0, false, false, err
		}
	}
	return kind, orderedEnum, casefoldText, nil
}

func (capabilities FieldCapabilities) validFor(fieldType FieldType) bool {
	kind, orderedEnum, casefoldText, err := fieldCapabilityTypeShape(fieldType)
	return err == nil && capabilities.Validate() == nil &&
		capabilities.logicalType == kind && capabilities.orderedEnum == orderedEnum &&
		capabilities.casefoldText == casefoldText
}
