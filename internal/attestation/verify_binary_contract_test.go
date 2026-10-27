package attestation

import (
	"bytes"
	"testing"
)

func TestVerifyBinarySchemaContractRejectsNonCanonicalAndUnknownJSON(t *testing.T) {
	release, migration, extension := binaryContractInputs(t)
	artifact, err := BuildBinarySchemaContract(release, migration, extension)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyBinarySchemaContract(append(artifact.CanonicalJSON(), '\n')); err == nil {
		t.Fatal("non-canonical contract was accepted")
	}
	unknown := bytes.Replace(
		artifact.CanonicalJSON(),
		[]byte(`{"component"`),
		[]byte(`{"another":true,"component"`),
		1,
	)
	if _, err := VerifyBinarySchemaContract(unknown); err == nil {
		t.Fatal("unknown contract member was accepted")
	}
}

func TestVerifiedBinarySchemaContractOutputsAreDefensive(t *testing.T) {
	release, migration, extension := binaryContractInputs(t)
	artifact, err := BuildBinarySchemaContract(release, migration, extension)
	if err != nil {
		t.Fatal(err)
	}
	first := artifact.CanonicalJSON()
	want := append([]byte(nil), first...)
	first[0] = 'x'
	if !bytes.Equal(artifact.CanonicalJSON(), want) {
		t.Fatal("verified binary contract bytes were mutated")
	}
	if (VerifiedBinarySchemaContract{}).CanonicalJSON() != nil ||
		(VerifiedBinarySchemaContract{}).DigestSHA256() != "" {
		t.Fatal("unverified binary contract exposed output")
	}
}
