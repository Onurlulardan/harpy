package publication

import (
	"bytes"
	"time"
)

type Release struct {
	id              ID
	resourceVersion [16]byte
	audience        Audience
	state           State
	cachePolicy     CachePolicy
	manifestDigest  [32]byte
	requester       [16]byte
	approver        [16]byte
	revision        uint64
}

func NewRelease(
	id ID, resourceVersion [16]byte, audience Audience, cachePolicy CachePolicy,
	manifestDigest [32]byte, requester [16]byte,
) (Release, error) {
	if !id.Valid() || !validV7(resourceVersion) || !audience.Valid() ||
		!cachePolicy.Valid() || zeroDigest(manifestDigest) || !validPrincipal(requester) {
		return Release{}, ErrInvalid
	}
	return Release{id: id, resourceVersion: resourceVersion, audience: audience,
		state: StateDraft, cachePolicy: cachePolicy, manifestDigest: manifestDigest,
		requester: requester, revision: 1}, nil
}

func (value Release) ID() ID                   { return value.id }
func (value Release) Audience() Audience       { return value.audience }
func (value Release) State() State             { return value.state }
func (value Release) CachePolicy() CachePolicy { return value.cachePolicy }
func (value Release) Revision() uint64         { return value.revision }
func (value Release) Approver() [16]byte       { return value.approver }

func (value Release) Submit(expected uint64) (Release, error) {
	if value.state != StateDraft || expected == 0 || expected != value.revision {
		return Release{}, ErrConflict
	}
	value.state = StatePendingApproval
	value.revision++
	return value, nil
}

func (value Release) Approve(
	expected uint64, approver [16]byte, stepUpAt, now time.Time,
) (Release, error) {
	if value.state != StatePendingApproval || expected == 0 || expected != value.revision ||
		!validPrincipal(approver) || approver == value.requester || now.IsZero() ||
		stepUpAt.IsZero() || stepUpAt.After(now) || now.Sub(stepUpAt) > 10*time.Minute {
		return Release{}, ErrConflict
	}
	value.approver = approver
	value.revision++
	return value, nil
}

func (value Release) Activate(expected uint64) (Release, error) {
	if value.state != StatePendingApproval || expected != value.revision ||
		(value.audience != AudienceTenant && !validPrincipal(value.approver)) {
		return Release{}, ErrConflict
	}
	value.state = StateActive
	value.revision++
	return value, nil
}

func (value Release) Revoke(expected uint64) (Release, error) {
	if value.state != StateActive || expected != value.revision {
		return Release{}, ErrConflict
	}
	value.state = StateRevoked
	value.revision++
	return value, nil
}

func validPrincipal(value [16]byte) bool { var zero [16]byte; return value != zero }
func zeroDigest(value [32]byte) bool     { return bytes.Equal(value[:], make([]byte, 32)) }
