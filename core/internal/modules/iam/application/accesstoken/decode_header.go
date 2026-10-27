package accesstoken

func decodeHeader(document []byte) (string, bool) {
	fields, ok := decodeStrictObject(document)
	if !ok || len(fields) != 3 {
		return "", false
	}
	typeValue, typeOK := decodeString(fields["typ"])
	algorithm, algorithmOK := decodeString(fields["alg"])
	keyID, keyOK := decodeString(fields["kid"])
	if !typeOK || !algorithmOK || !keyOK ||
		typeValue != Type || algorithm != Algorithm || !validKeyID(keyID) {
		return "", false
	}
	return keyID, true
}
