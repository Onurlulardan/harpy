package recordkernel

func physicalCoverageV2Valid(catalog Catalog, handles []physicalHandleDocumentV2) bool {
	base := make([]physicalHandleDocument, len(handles))
	for index, handle := range handles {
		base[index] = physicalHandleDocument{LogicalID: handle.LogicalID,
			FieldID: handle.FieldID, Kind: handle.Kind, PhysicalName: handle.PhysicalName,
			StorageType: handle.StorageType, Variant: handle.Variant,
			Generation: handle.Generation, Profile: handle.Profile}
	}
	return physicalCoverageValid(catalog, catalog.Fields(), base)
}
