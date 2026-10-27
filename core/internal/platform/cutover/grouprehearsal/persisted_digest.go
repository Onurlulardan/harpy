package grouprehearsal

import (
	"crypto/sha256"
	"encoding/binary"
)

const persistedDigestDomain = "harpy/cutover/billing-forms-rehearsal/v2\x00"

func digestPersisted(value PersistedRehearsal) [sha256.Size]byte {
	hash := sha256.New()
	_, _ = hash.Write([]byte(persistedDigestDomain))
	_, _ = hash.Write(value.Evidence.TenantID[:])
	number := make([]byte, 8)
	binary.BigEndian.PutUint64(number, uint64(value.Evidence.Generation))
	_, _ = hash.Write(number)
	for _, member := range value.Evidence.Members {
		_, _ = hash.Write([]byte(member.Domain))
		_, _ = hash.Write(member.OperationID[:])
		_, _ = hash.Write(member.RunID[:])
		_, _ = hash.Write(member.RollbackReceiptID[:])
		binary.BigEndian.PutUint64(number, member.SourceCount)
		_, _ = hash.Write(number)
		binary.BigEndian.PutUint64(number, member.TargetCount)
		_, _ = hash.Write(number)
		binary.BigEndian.PutUint64(number, member.DifferenceCount)
		_, _ = hash.Write(number)
		_, _ = hash.Write(member.AcceptanceSHA256[:])
		_, _ = hash.Write(member.RollbackSHA256[:])
	}
	_, _ = hash.Write([]byte{1, 1, 0})
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result
}
