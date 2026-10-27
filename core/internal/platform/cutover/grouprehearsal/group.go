package grouprehearsal

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/google/uuid"
)

const (
	MandatoryGroupID  = "billing+forms"
	CanonicalGroupID  = "billing_forms"
	WriterV1          = "v1"
	groupDigestDomain = "harpy/cutover/billing-forms-rehearsal/v1\x00"
)

type MemberEvidence struct {
	Domain            string
	OperationID       uuid.UUID
	RunID             uuid.UUID
	RollbackReceiptID uuid.UUID
	SourceCount       uint64
	TargetCount       uint64
	DifferenceCount   uint64
	AcceptanceSHA256  [sha256.Size]byte
	RollbackSHA256    [sha256.Size]byte
	WriterBefore      string
	WriterAfter       string
}

type GroupEvidence struct {
	TenantID            uuid.UUID
	Generation          int64
	GroupID             string
	Members             []MemberEvidence
	QuiesceDrained      bool
	UsageFlushed        bool
	ProductionDualWrite bool
}

type Acceptance struct{ digest [sha256.Size]byte }

func EvaluateGroup(value GroupEvidence) (Acceptance, error) {
	if value.TenantID == uuid.Nil || value.Generation < 1 || value.GroupID != MandatoryGroupID ||
		len(value.Members) != 1 || !value.QuiesceDrained || !value.UsageFlushed ||
		value.ProductionDualWrite {
		return Acceptance{}, ErrInvalid
	}
	want := []string{"billing"}
	for index, member := range value.Members {
		if member.Domain != want[index] || member.SourceCount != member.TargetCount ||
			member.DifferenceCount != 0 || member.AcceptanceSHA256 == [sha256.Size]byte{} ||
			member.RollbackSHA256 == [sha256.Size]byte{} || member.WriterBefore != WriterV1 ||
			member.WriterAfter != WriterV1 {
			return Acceptance{}, ErrInvalid
		}
	}
	digest := digestGroup(value)
	if digest == [sha256.Size]byte{} {
		return Acceptance{}, ErrInvalid
	}
	return Acceptance{digest: digest}, nil
}

func (value Acceptance) SHA256() ([sha256.Size]byte, error) {
	if value.digest == [sha256.Size]byte{} {
		return [sha256.Size]byte{}, ErrInvalid
	}
	return value.digest, nil
}

func digestGroup(value GroupEvidence) [sha256.Size]byte {
	hash := sha256.New()
	_, _ = hash.Write([]byte(groupDigestDomain))
	_, _ = hash.Write(value.TenantID[:])
	number := make([]byte, 8)
	binary.BigEndian.PutUint64(number, uint64(value.Generation))
	_, _ = hash.Write(number)
	for _, member := range value.Members {
		_, _ = hash.Write([]byte(member.Domain))
		binary.BigEndian.PutUint64(number, member.SourceCount)
		_, _ = hash.Write(number)
		_, _ = hash.Write(member.AcceptanceSHA256[:])
		_, _ = hash.Write(member.RollbackSHA256[:])
	}
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func (Acceptance) String() string   { return "[REDACTED]" }
func (Acceptance) GoString() string { return "[REDACTED]" }
func (Acceptance) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cutover group acceptance is not serializable")
}
