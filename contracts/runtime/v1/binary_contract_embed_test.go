package runtimev1

import (
	"errors"
	"testing"
	"testing/fstest"
)

func TestReadBinarySchemaContractFailsClosedWhenInstanceIsAbsent(t *testing.T) {
	files := fstest.MapFS{
		"binary-schema-contract.schema.json": &fstest.MapFile{Data: []byte(`{"type":"object"}`)},
	}
	_, err := readBinarySchemaContract(files)
	var missing BinarySchemaContractNotEmbeddedError
	if !errors.As(err, &missing) {
		t.Fatalf("expected missing embedded contract error, got %v", err)
	}
}

func TestReadBinarySchemaContractReturnsDefensiveBytes(t *testing.T) {
	want := []byte(`{"contract_version":"database-schema-contract.v1"}`)
	files := fstest.MapFS{
		"binary-schema-contract.json": &fstest.MapFile{Data: append([]byte(nil), want...)},
	}
	first, err := readBinarySchemaContract(files)
	if err != nil {
		t.Fatalf("read contract: %v", err)
	}
	first[0] = 'x'
	second, err := readBinarySchemaContract(files)
	if err != nil {
		t.Fatalf("read contract again: %v", err)
	}
	if string(second) != string(want) {
		t.Fatalf("embedded contract was mutated: %q", second)
	}
}
