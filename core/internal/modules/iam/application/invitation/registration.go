package invitation

import (
	"bytes"
	"crypto/sha256"

	"github.com/google/uuid"
)

type RegistrationProfile struct {
	SchemaID uuid.UUID
	Values   []byte
	SHA256   [32]byte
}

func (profile RegistrationProfile) Validate() error {
	if profile.SchemaID == uuid.Nil || len(profile.Values) < 2 ||
		len(profile.Values) > 1<<20 || profile.SHA256 == [32]byte{} ||
		sha256.Sum256(profile.Values) != profile.SHA256 {
		return ErrInvalid
	}
	return nil
}

type RegistrationPassword struct {
	Salt [16]byte
	Tag  [32]byte
}

func (credential RegistrationPassword) Validate() error {
	if credential.Salt == [16]byte{} || credential.Tag == [32]byte{} {
		return ErrInvalid
	}
	return nil
}

type DirectoryProjection struct {
	LookupKeyID  string
	HMAC         [32]byte
	SourceSHA256 [32]byte
}

func (projection DirectoryProjection) Validate() error {
	if !validTokenKeyID(projection.LookupKeyID) || projection.HMAC == [32]byte{} ||
		projection.SourceSHA256 == [32]byte{} {
		return ErrInvalid
	}
	return nil
}

func cloneRegistrationProfile(profile RegistrationProfile) RegistrationProfile {
	profile.Values = bytes.Clone(profile.Values)
	return profile
}
