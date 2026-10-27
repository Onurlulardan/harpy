package licenseauthority

import "github.com/google/uuid"

type SequenceDecision string

const (
	SequenceAccept SequenceDecision = "accept"
	SequenceNoOp   SequenceDecision = "noop"
)

type AcceptedHead struct {
	LicenseID, EnvironmentID uuid.UUID
	Sequence                 int64
	Digest                   string
}

func SequenceDecisionFor(
	current *AcceptedHead,
	candidate VerifiedSnapshot,
) (SequenceDecision, error) {
	if current == nil {
		return SequenceAccept, nil
	}
	if current.LicenseID != candidate.LicenseID() ||
		current.EnvironmentID != candidate.EnvironmentID() || current.Sequence < 1 ||
		!digestPattern.MatchString(current.Digest) {
		return "", ErrWrongBinding
	}
	if candidate.Sequence() < current.Sequence {
		return "", ErrSequenceRollback
	}
	if candidate.Sequence() == current.Sequence {
		if candidate.Digest() != current.Digest {
			return "", ErrSequenceConflict
		}
		return SequenceNoOp, nil
	}
	return SequenceAccept, nil
}
