package passwordenrollment

import (
	"context"
	"errors"
	"reflect"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

var (
	ErrInvalidCommand = errors.New("IAM password enrollment command is invalid")
	ErrStorage        = errors.New("IAM password enrollment storage failed")
)

type TenantScope = platformtx.TenantScope
type MutationEvidence = identitydomain.MutationEvidence
type MutationResult = identitydomain.MutationResult
type IdentityID = identitydomain.IdentityID

type Credential struct {
	salt [16]byte
	tag  [32]byte
}

func NewCredential(salt [16]byte, tag [32]byte) (Credential, error) {
	if salt == [16]byte{} || tag == [32]byte{} {
		return Credential{}, ErrInvalidCommand
	}
	return Credential{salt: salt, tag: tag}, nil
}

func (credential Credential) Valid() bool {
	return credential.salt != [16]byte{} && credential.tag != [32]byte{}
}

func (credential Credential) Salt() [16]byte { return credential.salt }
func (credential Credential) Tag() [32]byte  { return credential.tag }

type Command struct {
	UserID     IdentityID
	Credential Credential
	Evidence   MutationEvidence
}

func (command Command) Validate() error {
	if command.UserID == (IdentityID{}) || !command.Credential.Valid() ||
		command.Evidence.Validate() != nil {
		return ErrInvalidCommand
	}
	return nil
}

type Store interface {
	Create(context.Context, TenantScope, Command) (MutationResult, error)
}

func nilInterface(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	return reflected.Kind() == reflect.Pointer && reflected.IsNil()
}
