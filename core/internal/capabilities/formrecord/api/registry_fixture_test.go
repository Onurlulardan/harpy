package recordkernel

func fixtureRegistrySnapshot() RegistrySnapshot {
	types := knownTypeIDs()
	descriptors := make([]Descriptor, len(types))
	for index, typeID := range types {
		profile, _ := expectedValueProfile(typeID)
		roles := []Role{RoleStored}
		if typeID == TypeRelation {
			roles = []Role{RoleRelation}
		} else if expressionTypeAllowed(typeID) {
			roles = []Role{RoleStored, RoleCalculation}
		}
		descriptors[index] = Descriptor{TypeID: typeID,
			ValueProfile: profile, Roles: roles,
			Storage: StorageCapability{AllowedFeatures: []StorageFeature{}},
			Query: QueryCapability{Operators: []OperatorCapability{},
				Aggregates: []AggregateCapability{}},
			Cost: CostProfile{Dimensions: []BudgetDimension{}},
		}
	}
	return RegistrySnapshot{ContractVersion: RegistryContractVersion,
		HardLimits: expectedHardLimits(), Types: descriptors}
}

func expressionTypeAllowed(typeID TypeID) bool {
	return typeID == TypeText || typeID == TypeInteger || typeID == TypeDecimal ||
		typeID == TypeBoolean || typeID == TypeDate || typeID == TypeTimestamp ||
		typeID == TypeTime
}
