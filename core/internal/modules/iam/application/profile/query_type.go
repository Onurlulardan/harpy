package profile

import restquery "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func queryLogicalType(field Field) (restquery.LogicalType, error) {
	if field.cardinality == CardinalityList {
		switch field.kind {
		case LogicalString:
			return restquery.LogicalTypeTextList, nil
		case LogicalUUID:
			return restquery.LogicalTypeUUIDList, nil
		default:
			return 0, ErrQueryCapabilityRejected
		}
	}
	switch field.kind {
	case LogicalString:
		return restquery.LogicalTypeText, nil
	case LogicalInteger:
		return restquery.LogicalTypeInteger, nil
	case LogicalDecimal:
		return restquery.LogicalTypeDecimal, nil
	case LogicalBoolean:
		return restquery.LogicalTypeBoolean, nil
	case LogicalDate:
		return restquery.LogicalTypeDate, nil
	case LogicalTimestamp:
		return restquery.LogicalTypeTimestamp, nil
	case LogicalUUID:
		return restquery.LogicalTypeUUID, nil
	case LogicalEnum:
		return restquery.LogicalTypeEnum, nil
	case LogicalJSON:
		return restquery.LogicalTypeJSON, nil
	default:
		return 0, ErrQueryCapabilityRejected
	}
}
