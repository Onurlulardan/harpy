package recordkernel

func materializeStorageShape(catalog Catalog, record *Record) {
	for _, field := range catalog.fields {
		value, present := record.values[field.ID]
		if relationMany(field) && (!present || value.null) {
			record.values[field.ID] = typedValue(TypeRelation, []byte("[]"))
			continue
		}
		if !present {
			record.values[field.ID] = nullValue(field.TypeID)
		}
	}
}

func recordStorageNormalized(catalog Catalog, record Record) bool {
	if len(record.values) != len(catalog.fields) {
		return false
	}
	for _, field := range catalog.fields {
		value, present := record.values[field.ID]
		if !present || relationMany(field) && value.null {
			return false
		}
	}
	return true
}

func relationMany(field Field) bool {
	return field.TypeID == TypeRelation && field.Relation != nil &&
		field.Relation.Cardinality == "many"
}
