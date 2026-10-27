package systemtransfer

import "sync/atomic"

type sourceCleanupAdmissionStub struct {
	open atomic.Bool
}

func (gate *sourceCleanupAdmissionStub) Accepting() bool {
	return gate != nil && gate.open.Load()
}

func (gate *sourceCleanupAdmissionStub) Open() bool {
	if gate == nil {
		return false
	}
	gate.open.Store(true)
	return true
}
