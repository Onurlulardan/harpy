package api

import (
	"io"

	fcartv1 "github.com/Onurlulardan/harpy/contracts/fcart/v1"
)

type ArtifactHeader = fcartv1.Header
type ArtifactMetadata = fcartv1.CiphertextMetadata
type DataEncryptionKey = fcartv1.DEK
type KeyEnvelope = fcartv1.KeyEnvelope
type KeyIdentity = fcartv1.KeyIdentity
type EncryptResult = fcartv1.EncryptResult
type DecryptResult = fcartv1.DecryptResult

func NewArtifactHeader(identity KeyIdentity, format Format) (ArtifactHeader, error) {
	return fcartv1.NewHeader(identity.ArtifactID, identity.AttemptGeneration,
		fcartFormat(format))
}

func Encrypt(destination io.Writer, plaintext io.Reader, header ArtifactHeader,
	key DataEncryptionKey) (EncryptResult, error) {
	return fcartv1.Encrypt(destination, plaintext, header, key,
		fcartv1.NewCryptographicRandomSource())
}

func Decrypt(destination io.Writer, ciphertext io.Reader,
	metadata ArtifactMetadata, key DataEncryptionKey) (DecryptResult, error) {
	return fcartv1.Decrypt(destination, ciphertext, metadata, key)
}

func fcartFormat(format Format) fcartv1.Format {
	switch format {
	case FormatCSV:
		return fcartv1.FormatCSV
	case FormatJSONL:
		return fcartv1.FormatJSONL
	case FormatXLSX:
		return fcartv1.FormatXLSX
	default:
		return ""
	}
}
