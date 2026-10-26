package fcartv1

import (
	"encoding/hex"
	"testing"
)

func TestEncryptDecryptGoldenArtifact(t *testing.T) {
	plaintext := []byte("id,name\r\n1,Ada\r\n")
	ciphertext, encrypted := encryptFixture(t, plaintext)
	const goldenHex = "46434152543100112233445566778899aabbccddeeff0000000000000007" +
		"0100800000000000000000000000000010a500000000000000000000004a" +
		"56d1c3769445640b3e6c175d90ba138819f50e119400df34700063e72b57" +
		"ec"
	if actual := hex.EncodeToString(ciphertext); actual != goldenHex {
		t.Fatalf("FCART1 golden bytes changed:\n%s", actual)
	}
	if encrypted.Header != fixtureHeader(t) ||
		encrypted.CiphertextBytes != uint64(len(ciphertext)) ||
		encrypted.PlaintextBytes != uint64(len(plaintext)) ||
		encrypted.ChunkCount != 1 {
		t.Fatalf("unexpected encryption result: %#v", encrypted)
	}
	decoded, decrypted, err := decryptFixture(
		t, ciphertext, encrypted.CiphertextMetadata,
	)
	if err != nil {
		t.Fatal(err)
	}
	if string(decoded) != string(plaintext) {
		t.Fatalf("plaintext = %q", decoded)
	}
	if decrypted.PlaintextBytes != uint64(len(plaintext)) ||
		decrypted.ChunkCount != encrypted.ChunkCount {
		t.Fatalf("unexpected decryption result: %#v", decrypted)
	}
}

func TestEmptyPlaintextIsCanonicalHeaderOnlyArtifact(t *testing.T) {
	ciphertext, result := encryptFixture(t, nil)
	if len(ciphertext) != HeaderBytes || result.ChunkCount != 0 ||
		result.PlaintextBytes != 0 {
		t.Fatalf("empty artifact result = %#v, bytes = %d", result, len(ciphertext))
	}
	decoded, decrypted, err := decryptFixture(
		t, ciphertext, result.CiphertextMetadata,
	)
	if err != nil || len(decoded) != 0 || decrypted != (DecryptResult{}) {
		t.Fatalf("empty round trip = %x, %#v, %v", decoded, decrypted, err)
	}
}
