package resolved

func writeRecordQuery(
	writer *identityWriter, query RecordQuery, includeCursor bool,
) {
	writer.writeUint32(uint32(len(query.projections)))
	for _, projection := range query.projections {
		writeProjection(writer, projection)
	}
	writer.writeBool(query.where != nil)
	if query.where != nil {
		writeFilter(writer, *query.where)
	}
	writer.writeUint32(uint32(len(query.sorts)))
	for _, sort := range query.sorts {
		writeSort(writer, sort)
	}
	writeSearch(writer, query.search)
	writer.writeUint64(uint64(query.page.size))
	writer.writeBool(includeCursor && query.page.hasAfter)
	if includeCursor && query.page.hasAfter {
		writer.writeString(query.page.after)
	}
	writer.writeByte(byte(query.count))
	writeRequestPath(writer, query.countPath)
	writer.writeBool(query.exactCountEnabled)
	writer.writeUint32(uint32(len(query.includes)))
	for _, include := range query.includes {
		writeInclude(writer, include)
	}
}

func writeInclude(writer *identityWriter, include Include) {
	relationID, err := include.relation.RelationID()
	if err != nil {
		writer.err = err
		return
	}
	cardinality, err := include.relation.Cardinality()
	if err != nil {
		writer.err = err
		return
	}
	writer.writeBytes(relationID[:])
	writer.writeByte(byte(cardinality))
	writeCatalogReference(writer, include.target)
	writeResourceStorage(writer, include.targetStorage)
	writeRelationStorage(writer, include.relationStorage)
	writeField(writer, include.stableID)
	writeRequestPath(writer, include.path)
	writer.writeUint32(uint32(len(include.projections)))
	for _, projection := range include.projections {
		writeProjection(writer, projection)
	}
	writer.writeBool(include.where != nil)
	if include.where != nil {
		writeFilter(writer, *include.where)
	}
	writer.writeUint32(uint32(len(include.sorts)))
	for _, sort := range include.sorts {
		writeSort(writer, sort)
	}
	writer.writeUint64(uint64(include.limit))
	writer.writeUint32(uint32(len(include.children)))
	for _, child := range include.children {
		writeInclude(writer, child)
	}
}
