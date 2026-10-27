package identity

import (
	"bytes"
	"crypto/sha256"
	"sort"
)

const MaximumUserRelationSetSize = 100

type relationSetPatch struct {
	replace bool
	targets []IdentityID
	rowIDs  []IdentityID
}

func newRelationSetPatch(
	userID IdentityID,
	replace bool,
	targets []IdentityID,
	relationKind string,
) (relationSetPatch, error) {
	if !replace {
		if len(targets) != 0 {
			return relationSetPatch{}, ErrInvalidCommand
		}
		return relationSetPatch{}, nil
	}
	if userID == (IdentityID{}) || len(targets) > MaximumUserRelationSetSize {
		return relationSetPatch{}, ErrInvalidCommand
	}
	canonical := append([]IdentityID(nil), targets...)
	sort.Slice(canonical, func(left, right int) bool {
		return bytes.Compare(canonical[left][:], canonical[right][:]) < 0
	})
	rows := make([]IdentityID, len(canonical))
	for index, targetID := range canonical {
		if targetID == (IdentityID{}) ||
			index > 0 && targetID == canonical[index-1] {
			return relationSetPatch{}, ErrInvalidCommand
		}
		rows[index] = deterministicRelationRowID(relationKind, userID, targetID)
	}
	return relationSetPatch{replace: true, targets: canonical, rowIDs: rows}, nil
}

func (patch relationSetPatch) validate(
	userID IdentityID,
	relationKind string,
) error {
	if !patch.replace {
		if len(patch.targets) == 0 && len(patch.rowIDs) == 0 {
			return nil
		}
		return ErrInvalidCommand
	}
	if userID == (IdentityID{}) || len(patch.targets) > MaximumUserRelationSetSize ||
		len(patch.targets) != len(patch.rowIDs) {
		return ErrInvalidCommand
	}
	for index, targetID := range patch.targets {
		if targetID == (IdentityID{}) || patch.rowIDs[index] == (IdentityID{}) ||
			index > 0 && bytes.Compare(patch.targets[index-1][:], targetID[:]) >= 0 ||
			patch.rowIDs[index] != deterministicRelationRowID(relationKind, userID, targetID) {
			return ErrInvalidCommand
		}
	}
	return nil
}

func deterministicRelationRowID(
	relationKind string,
	userID IdentityID,
	targetID IdentityID,
) IdentityID {
	hasher := sha256.New()
	_, _ = hasher.Write([]byte("harpy/iam/user-relation-row/v1\x00"))
	_, _ = hasher.Write([]byte(relationKind))
	_, _ = hasher.Write([]byte{0})
	_, _ = hasher.Write(userID[:])
	_, _ = hasher.Write(targetID[:])
	digest := hasher.Sum(nil)
	var result IdentityID
	copy(result[:], digest[:len(result)])
	result[6] = result[6]&0x0f | 0x80
	result[8] = result[8]&0x3f | 0x80
	return result
}

func cloneIdentityIDs(values []IdentityID) []IdentityID {
	return append([]IdentityID(nil), values...)
}
