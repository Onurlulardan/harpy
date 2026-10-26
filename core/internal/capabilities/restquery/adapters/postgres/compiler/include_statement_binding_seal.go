package compiler

func writeIncludeStatementBinding(
	writer *sealWriter, binding includeStatementBinding,
) error {
	if writer == nil || binding.validate() != nil ||
		writeIncludePlanBinding(writer, binding.plan) != nil {
		return ErrContract
	}
	writer.writeBytes(binding.graphSeal[:])
	writer.writeBytes(binding.nodeSeal[:])
	writer.writeInt(int64(binding.role))
	return nil
}
