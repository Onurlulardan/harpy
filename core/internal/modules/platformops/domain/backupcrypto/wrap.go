package backupcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/hpke"
	"encoding/base64"
	"io"
)

func wrapDEK(context Context, dek []byte, keys SealKeys) (keyWrap, keyWrap, error) {
	if len(dek) != 32 || len(keys.DRKEK) != 32 || len(keys.RecoveryPublic) != 32 ||
		keys.Random == nil || !validContext(context) {
		return keyWrap{}, keyWrap{}, ErrInvalidArchive
	}
	drBlock, err := aes.NewCipher(keys.DRKEK)
	if err != nil {
		return keyWrap{}, keyWrap{}, ErrInvalidArchive
	}
	drAEAD, err := cipher.NewGCM(drBlock)
	if err != nil {
		return keyWrap{}, keyWrap{}, ErrInvalidArchive
	}
	nonce := make([]byte, drAEAD.NonceSize())
	if _, err = io.ReadFull(keys.Random, nonce); err != nil {
		return keyWrap{}, keyWrap{}, ErrInvalidArchive
	}
	drCiphertext := drAEAD.Seal(nil, nonce, dek, wrapInfo(context))
	public, err := ecdh.X25519().NewPublicKey(keys.RecoveryPublic)
	if err != nil {
		return keyWrap{}, keyWrap{}, ErrInvalidArchive
	}
	hpkePublic, err := hpke.NewDHKEMPublicKey(public)
	if err != nil {
		return keyWrap{}, keyWrap{}, ErrInvalidArchive
	}
	hpkeCiphertext, err := hpke.Seal(hpkePublic, hpke.HKDFSHA256(),
		hpke.AES256GCM(), wrapInfo(context), dek)
	if err != nil {
		return keyWrap{}, keyWrap{}, ErrInvalidArchive
	}
	return keyWrap{Algorithm: "AES-256-GCM", KeyID: context.DRKeyID,
			Nonce: encode(nonce), Ciphertext: encode(drCiphertext)},
		keyWrap{Algorithm: "HPKE-X25519-HKDF-SHA256-AES256GCM",
			KeyID: context.RecoveryKeyID, Ciphertext: encode(hpkeCiphertext)}, nil
}

func unwrapDEK(header archiveHeader, keys OpenKeys) ([]byte, error) {
	drSelected := len(keys.DRKEK) > 0
	hpkeSelected := len(keys.RecoveryPrivate) > 0
	if drSelected == hpkeSelected {
		return nil, ErrInvalidArchive
	}
	context := header.context()
	if drSelected {
		return unwrapDR(header.DRWrap, context, keys.DRKEK)
	}
	private, err := ecdh.X25519().NewPrivateKey(keys.RecoveryPrivate)
	if err != nil {
		return nil, ErrInvalidArchive
	}
	hpkePrivate, err := hpke.NewDHKEMPrivateKey(private)
	if err != nil {
		return nil, ErrInvalidArchive
	}
	ciphertext, err := decode(header.RecoveryWrap.Ciphertext)
	if err != nil {
		return nil, err
	}
	dek, err := hpke.Open(hpkePrivate, hpke.HKDFSHA256(), hpke.AES256GCM(),
		wrapInfo(context), ciphertext)
	if err != nil || len(dek) != 32 {
		return nil, ErrInvalidArchive
	}
	return dek, nil
}

func unwrapDR(value keyWrap, context Context, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, ErrInvalidArchive
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrInvalidArchive
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrInvalidArchive
	}
	nonce, nonceErr := decode(value.Nonce)
	ciphertext, ciphertextErr := decode(value.Ciphertext)
	if nonceErr != nil || ciphertextErr != nil {
		return nil, ErrInvalidArchive
	}
	dek, err := aead.Open(nil, nonce, ciphertext, wrapInfo(context))
	if err != nil || len(dek) != 32 {
		return nil, ErrInvalidArchive
	}
	return dek, nil
}

func encode(value []byte) string { return base64.RawURLEncoding.EncodeToString(value) }

func decode(value string) ([]byte, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, ErrInvalidArchive
	}
	return decoded, nil
}
