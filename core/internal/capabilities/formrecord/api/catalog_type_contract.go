package recordkernel

func expectedValueProfile(typeID TypeID) (ValueProfile, bool) {
	profiles := map[TypeID]ValueProfile{
		TypeText:            ValueTextNFCV1,
		TypeInteger:         ValueIntegerBase10V1,
		TypeDecimal:         ValueDecimalBase10V1,
		TypeBoolean:         ValueBooleanV1,
		TypeDate:            ValueDateISO8601V1,
		TypeTimestamp:       ValueTimestampUTCMicrosecondV1,
		TypeUUID:            ValueUUIDRFC4122LowerV1,
		TypeEnum:            ValueEnumCodeV1,
		TypeRelation:        ValueRelationUUIDV1,
		TypeTextList:        ValueTextListNFCV1,
		TypeUUIDList:        ValueUUIDListV1,
		TypeJSON:            ValueJSONRFC8785V1,
		TypeTime:            ValueTimeMicrosecondV1,
		TypeDocumentRef:     ValueDocumentRefUUIDV1,
		TypeDocumentRefList: ValueDocumentRefListV1,
		TypeEnumList:        ValueEnumListCodesV1,
	}
	value, found := profiles[typeID]
	return value, found
}

func catalogRoleAllowed(typeID TypeID, role Role) bool {
	if typeID == TypeRelation {
		return role == RoleRelation
	}
	switch typeID {
	case TypeText, TypeInteger, TypeDecimal,
		TypeBoolean, TypeDate, TypeTimestamp,
		TypeTime:
		return role == RoleStored || role == RoleCalculation
	default:
		return knownType(typeID) && role == RoleStored
	}
}
