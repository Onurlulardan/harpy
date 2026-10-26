package compiler

func writeIncludePlanBinding(
	writer *sealWriter, binding includePlanBinding,
) error {
	if writer == nil || binding.validate() != nil {
		return ErrContract
	}
	writer.writeBytes(binding.tenant[:])
	writer.writeBytes(binding.actor[:])
	writer.writeBytes(binding.authorization[:])
	writer.writeBytes(binding.mandatory[:])
	writer.writeBytes(binding.identity[:])
	return writeCatalogReference(writer, binding.rootCatalog)
}
