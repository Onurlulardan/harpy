package binarycontract

import runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"

func LoadEmbedded() (Artifact, error) {
	return loadEmbedded(runtimev1.ReadEmbeddedBinarySchemaContract)
}

func loadEmbedded(read func() ([]byte, error)) (Artifact, error) {
	document, err := read()
	if err != nil {
		return Artifact{}, err
	}
	return Verify(document)
}
