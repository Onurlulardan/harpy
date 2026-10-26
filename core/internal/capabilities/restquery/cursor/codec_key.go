package cursor

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type codecKey struct {
	kid       string
	aead      cipher.AEAD
	digestKey [sha256.Size]byte
}

func newCodecKey(key config.CursorKey) (codecKey, error) {
	kid, kidErr := key.KID()
	material, materialErr := key.AES256Key()
	if kidErr != nil || materialErr != nil {
		return codecKey{}, errInvalidCodec
	}
	block, blockErr := aes.NewCipher(material[:])
	derived, deriveErr := hkdf.Key(
		sha256.New, material[:], nil, mandatoryPredicateDigestInfo, sha256.Size,
	)
	clear(material[:])
	if blockErr != nil || deriveErr != nil || len(derived) != sha256.Size {
		clear(derived)
		return codecKey{}, errInvalidCodec
	}
	aead, aeadErr := cipher.NewGCM(block)
	if aeadErr != nil || aead.NonceSize() != nonceBytes || aead.Overhead() != 16 {
		clear(derived)
		return codecKey{}, errInvalidCodec
	}
	result := codecKey{kid: kid, aead: aead}
	copy(result.digestKey[:], derived)
	clear(derived)
	return result, nil
}

func (key codecKey) validate() error {
	if key.kid == "" || key.aead == nil || key.aead.NonceSize() != nonceBytes ||
		key.aead.Overhead() != 16 || key.digestKey == [sha256.Size]byte{} {
		return errInvalidCodec
	}
	return nil
}

func (codec Codec) decodeKey(kid string) (codecKey, bool) {
	if kid == codec.state.current.kid {
		return codec.state.current, true
	}
	if kid == codec.state.previous.kid {
		return codec.state.previous, true
	}
	return codecKey{}, false
}
