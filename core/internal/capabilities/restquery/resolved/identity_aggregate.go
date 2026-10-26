package resolved

func writeAggregateQuery(writer *identityWriter, query AggregateQuery) {
	writer.writeUint32(uint32(len(query.groups)))
	for _, group := range query.groups {
		writeField(writer, group.field)
		writeOutputKey(writer, group.key)
		writeRequestPath(writer, group.path)
	}
	writer.writeUint32(uint32(len(query.measures)))
	for _, measure := range query.measures {
		writeAggregateMeasure(writer, measure)
	}
	writer.writeBool(query.where != nil)
	if query.where != nil {
		writeFilter(writer, *query.where)
	}
	writeSearch(writer, query.search)
	writer.writeBool(query.having != nil)
	if query.having != nil {
		writeHaving(writer, *query.having)
	}
	writer.writeUint32(uint32(len(query.sorts)))
	for _, sort := range query.sorts {
		writeAggregateReference(writer, sort.reference)
		writer.writeByte(byte(sort.direction))
		writer.writeBool(sort.implicit)
	}
	writer.writeUint64(uint64(query.limit))
}

func writeAggregateMeasure(writer *identityWriter, measure AggregateMeasure) {
	writeOutputKey(writer, measure.key)
	writer.writeByte(byte(measure.function))
	writer.writeBool(measure.hasField)
	if measure.hasField {
		writeField(writer, measure.field)
	}
	writeRequestPath(writer, measure.path)
}

func writeHaving(writer *identityWriter, having Having) {
	writer.writeByte(byte(having.kind))
	if having.predicate != nil {
		writeAggregateReference(writer, having.predicate.reference)
		writer.writeByte(byte(having.predicate.operator))
		writer.writeUint32(uint32(len(having.predicate.values)))
		for _, value := range having.predicate.values {
			writeValue(writer, value)
		}
		return
	}
	writer.writeUint32(uint32(len(having.children)))
	for _, child := range having.children {
		writeHaving(writer, child)
	}
}

func writeAggregateReference(writer *identityWriter, reference AggregateReference) {
	writer.writeByte(byte(reference.kind))
	writer.writeUint64(uint64(reference.ordinal))
	writeRequestPath(writer, reference.path)
}
