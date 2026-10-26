package archive

func RequestDigest(request CreateRequest) ([32]byte, bool) {
	_, _, digest, valid := canonicalRequest(request)
	return digest, valid
}
