package platformiam

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

const EnvironmentScope = "environment"

type AuthorityAcknowledgement struct {
	OperationID, RegistryBindingID, ScopeID ID
	ScopeKind                               string
	RegistrySequence, Generation            int64
	ProposalSHA256, PreviousHeadSHA256      [32]byte
	StateSHA256, AckSHA256                  [32]byte
	IssuedAt                                time.Time
	RegistryKID                             string
	Signature                               []byte
}

func (ack AuthorityAcknowledgement) Validate(now time.Time) error {
	expectedDigest := sha256.Sum256(ack.SigningBytes())
	if !ack.OperationID.Valid() || !ack.RegistryBindingID.Valid() || !ack.ScopeID.Valid() ||
		ack.ScopeKind != EnvironmentScope || ack.RegistrySequence < 1 || ack.Generation < 1 ||
		ack.ProposalSHA256 == [32]byte{} || ack.PreviousHeadSHA256 == [32]byte{} ||
		ack.StateSHA256 == [32]byte{} || ack.AckSHA256 != expectedDigest || ack.IssuedAt.IsZero() ||
		ack.IssuedAt.After(now) || !validKID(ack.RegistryKID) || len(ack.Signature) != ed25519.SignatureSize {
		return ErrInvalidCommand
	}
	return nil
}

func (ack AuthorityAcknowledgement) Verify(now time.Time, key ed25519.PublicKey) error {
	if ack.Validate(now) != nil || len(key) != ed25519.PublicKeySize ||
		!ed25519.Verify(key, ack.SigningBytes(), ack.Signature) {
		return ErrDenied
	}
	return nil
}

func (ack AuthorityAcknowledgement) SigningBytes() []byte {
	return []byte(strings.Join([]string{
		"protected-authority-ack.v1", ack.RegistryBindingID.String(), ack.ScopeKind,
		ack.ScopeID.String(), strconv.FormatInt(ack.RegistrySequence, 10),
		ack.OperationID.String(), hex.EncodeToString(ack.ProposalSHA256[:]),
		strconv.FormatInt(ack.Generation, 10), hex.EncodeToString(ack.PreviousHeadSHA256[:]),
		hex.EncodeToString(ack.StateSHA256[:]), ack.IssuedAt.UTC().Format(time.RFC3339Nano),
		ack.RegistryKID,
	}, "\n"))
}
