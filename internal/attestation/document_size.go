package attestation

func withinDocumentSizeLimit(size, maximumBytes int) bool {
	return maximumBytes > 0 && size > 0 && size <= maximumBytes
}
