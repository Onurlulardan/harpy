package resolved

func encodeQuerySpec(spec QuerySpec, includeCursor bool) ([]byte, error) {
	writer := &identityWriter{}
	writer.writeBytes([]byte{'R', 'Q', '2', 1})
	writer.writeByte(byte(spec.mode))
	writeCatalogReference(writer, spec.reference)
	writeResourceStorage(writer, spec.storage)
	writeField(writer, spec.stableID)
	writeRequestPath(writer, spec.rootPath)
	if spec.record != nil {
		writer.writeByte(1)
		writeRecordQuery(writer, *spec.record, includeCursor)
	} else {
		writer.writeByte(2)
		writeAggregateQuery(writer, *spec.aggregate)
	}
	return writer.result()
}

func writeFilter(writer *identityWriter, filter Filter) {
	writer.writeByte(byte(filter.kind))
	if filter.predicate != nil {
		writeField(writer, filter.predicate.field)
		writer.writeByte(byte(filter.predicate.operator))
		writeRequestPath(writer, filter.predicate.path)
		writer.writeUint32(uint32(len(filter.predicate.values)))
		for _, value := range filter.predicate.values {
			writeValue(writer, value)
		}
		return
	}
	writer.writeUint32(uint32(len(filter.children)))
	for _, child := range filter.children {
		writeFilter(writer, child)
	}
}

func writeSearch(writer *identityWriter, terms []SearchTerm) {
	writer.writeUint32(uint32(len(terms)))
	for _, term := range terms {
		writeField(writer, term.field)
		writeRequestPath(writer, term.path)
		writeValue(writer, term.value)
	}
}
