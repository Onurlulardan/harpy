package password

import (
	"errors"
	"testing"
)

func TestArgon2idCredentialFromStorageAcceptsOnlyV1Profile(t *testing.T) {
	salt := make([]byte, SaltBytesV1)
	tag := make([]byte, TagBytesV1)
	credential, err := Argon2idCredentialFromStorage(
		salt, tag, MemoryKiBV1, IterationsV1, ParallelismV1,
	)
	if err != nil || !credential.Valid() {
		t.Fatalf("valid profile rejected: %v", err)
	}
	tests := []struct {
		salt []byte
		tag  []byte
		m    uint32
		time uint32
		p    uint8
	}{
		{salt[:15], tag, MemoryKiBV1, IterationsV1, ParallelismV1},
		{salt, tag[:31], MemoryKiBV1, IterationsV1, ParallelismV1},
		{salt, tag, MemoryKiBV1 + 1, IterationsV1, ParallelismV1},
		{salt, tag, MemoryKiBV1, IterationsV1 + 1, ParallelismV1},
		{salt, tag, MemoryKiBV1, IterationsV1, ParallelismV1 + 1},
	}
	for index, test := range tests {
		if _, err := Argon2idCredentialFromStorage(
			test.salt, test.tag, test.m, test.time, test.p,
		); !errors.Is(err, ErrCredentialInvalid) {
			t.Fatalf("invalid profile %d error = %v", index, err)
		}
	}
	if !(Argon2idCredential{}).NeedsRehash() {
		t.Fatal("zero credential did not require replacement")
	}
}
