package fcartv1

import (
	"bytes"
	"testing"
)

func FuzzRoundTrip(f *testing.F) {
	f.Add([]byte(nil), uint8(0))
	f.Add([]byte("id,name\r\n1,Ada\r\n"), uint8(1))
	f.Add([]byte{0, 1, 2, 0xff}, uint8(2))
	f.Fuzz(func(t *testing.T, plaintext []byte, formatIndex uint8) {
		if len(plaintext) > 64*1024 {
			t.Skip()
		}
		formats := []Format{FormatCSV, FormatJSONL, FormatXLSX, FormatPDF}
		header, err := NewHeader(
			fixtureArtifactID, uint64(formatIndex)+1,
			formats[int(formatIndex)%len(formats)],
		)
		if err != nil {
			t.Fatal(err)
		}
		var ciphertext bytes.Buffer
		encrypted, err := Encrypt(
			&ciphertext, bytes.NewReader(plaintext), header, fixtureDEK(t),
			&deterministicNonceSource{seed: formatIndex},
		)
		if err != nil {
			t.Fatal(err)
		}
		var decoded bytes.Buffer
		decrypted, err := Decrypt(
			&decoded, bytes.NewReader(ciphertext.Bytes()),
			encrypted.CiphertextMetadata, fixtureDEK(t),
		)
		if err != nil || !bytes.Equal(decoded.Bytes(), plaintext) {
			t.Fatalf("round trip = %x, %v", decoded.Bytes(), err)
		}
		if decrypted.PlaintextBytes != uint64(len(plaintext)) ||
			decrypted.ChunkCount != encrypted.ChunkCount {
			t.Fatalf("counts = %#v / %#v", encrypted, decrypted)
		}
		wantCiphertextBytes := uint64(HeaderBytes) + uint64(len(plaintext)) +
			encrypted.ChunkCount*ChunkFramingBytes
		if encrypted.CiphertextBytes != wantCiphertextBytes {
			t.Fatalf("ciphertext bytes = %d, want %d", encrypted.CiphertextBytes, wantCiphertextBytes)
		}
	})
}
