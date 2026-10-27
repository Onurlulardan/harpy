package runtimev1

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
)

type BinarySchemaContractNotEmbeddedError struct{}

func (BinarySchemaContractNotEmbeddedError) Error() string {
	return "binary schema contract is not embedded: expected binary-schema-contract.json"
}

// The schema is the permanent match that keeps ordinary checkout builds valid.
// Release staging adds binary-schema-contract.json before compiling the binary.
//
//go:embed binary-schema-contract*.json
var binarySchemaContractFiles embed.FS

func ReadEmbeddedBinarySchemaContract() ([]byte, error) {
	return readBinarySchemaContract(binarySchemaContractFiles)
}

func readBinarySchemaContract(files fs.FS) ([]byte, error) {
	document, err := fs.ReadFile(files, "binary-schema-contract.json")
	if errors.Is(err, fs.ErrNotExist) {
		return nil, BinarySchemaContractNotEmbeddedError{}
	}
	if err != nil {
		return nil, fmt.Errorf("read embedded binary schema contract: %w", err)
	}
	return append([]byte(nil), document...), nil
}
