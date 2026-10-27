package fcartv1

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"testing"

	"github.com/google/uuid"
)

var fixtureArtifactID = uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")

type deterministicNonceSource struct {
	seed     byte
	sequence uint64
	fixed    *[NonceBytes]byte
	err      error
}

func (source *deterministicNonceSource) Nonce() ([NonceBytes]byte, error) {
	if source.err != nil {
		return [NonceBytes]byte{}, source.err
	}
	if source.fixed != nil {
		return *source.fixed, nil
	}
	var nonce [NonceBytes]byte
	nonce[0] = source.seed
	binary.BigEndian.PutUint64(nonce[4:], source.sequence)
	source.sequence++
	return nonce, nil
}

func (*deterministicNonceSource) isFCART1RandomSource() {}

func fixtureHeader(t testing.TB) Header {
	t.Helper()
	header, err := NewHeader(fixtureArtifactID, 7, FormatCSV)
	if err != nil {
		t.Fatal(err)
	}
	return header
}

func fixtureDEK(t testing.TB) DEK {
	t.Helper()
	raw := make([]byte, DEKBytes)
	for index := range raw {
		raw[index] = byte(index + 1)
	}
	key, err := NewDEK(raw)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func encryptFixture(t testing.TB, plaintext []byte) ([]byte, EncryptResult) {
	t.Helper()
	var ciphertext bytes.Buffer
	result, err := Encrypt(
		&ciphertext, bytes.NewReader(plaintext), fixtureHeader(t), fixtureDEK(t),
		&deterministicNonceSource{seed: 0xA5},
	)
	if err != nil {
		t.Fatal(err)
	}
	return ciphertext.Bytes(), result
}

func metadataForBytes(header Header, ciphertext []byte) CiphertextMetadata {
	encoded, _ := header.encode()
	return CiphertextMetadata{
		Header: header, HeaderSHA256: sha256.Sum256(encoded[:]),
		CiphertextBytes:  uint64(len(ciphertext)),
		CiphertextSHA256: sha256.Sum256(ciphertext),
	}
}

func cloneBytes(value []byte) []byte {
	return append([]byte(nil), value...)
}

func decryptFixture(
	t testing.TB,
	ciphertext []byte,
	metadata CiphertextMetadata,
) ([]byte, DecryptResult, error) {
	t.Helper()
	var plaintext bytes.Buffer
	result, err := Decrypt(
		&plaintext, bytes.NewReader(ciphertext), metadata, fixtureDEK(t),
	)
	return plaintext.Bytes(), result, err
}
