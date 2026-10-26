package recordkernel

type RegistryEvidence struct {
	Snapshot RegistrySnapshot
}

type registryCatalog struct {
	limits BudgetLimits
	types  map[TypeID]Descriptor
}

func loadRegistry(evidence RegistryEvidence) (registryCatalog, error) {
	if evidence.Snapshot.ContractVersion != RegistryContractVersion ||
		evidence.Snapshot.HardLimits != expectedHardLimits() ||
		len(evidence.Snapshot.Types) != len(knownTypeIDs()) {
		return registryCatalog{}, ErrCatalogInvalid
	}
	types := make(map[TypeID]Descriptor,
		len(evidence.Snapshot.Types))
	for _, descriptor := range evidence.Snapshot.Types {
		if !knownType(descriptor.TypeID) || descriptor.ValueProfile == "" ||
			len(descriptor.Roles) == 0 {
			return registryCatalog{}, ErrCatalogInvalid
		}
		if _, duplicate := types[descriptor.TypeID]; duplicate {
			return registryCatalog{}, ErrCatalogInvalid
		}
		types[descriptor.TypeID] = descriptor
	}
	for _, typeID := range knownTypeIDs() {
		if _, found := types[typeID]; !found {
			return registryCatalog{}, ErrCatalogInvalid
		}
	}
	return registryCatalog{limits: evidence.Snapshot.HardLimits, types: types}, nil
}
