package catalogcomponentv1

import "time"

type Component struct {
	ID            string
	Owner         string
	Generation    int64
	Contract      string
	ContentSHA256 string
}

type Readiness struct {
	Component
	WriterV2 bool
	State    string
	ReadyAt  time.Time
}

const (
	ProvisionAvailable = "AVAILABLE"
	ProvisionStaged    = "STAGED"
)

type ProvisionRequest struct {
	PackageID      string
	PackageSlug    string
	PackageVersion string
	ReleaseID      string
	Component      Component
	Document       []byte
	IdempotencyKey string
}

func (request ProvisionRequest) Clone() ProvisionRequest {
	request.Document = append([]byte(nil), request.Document...)
	return request
}

type ProvisionReceipt struct {
	ID             string
	PackageID      string
	PackageSlug    string
	PackageVersion string
	ReleaseID      string
	Component      Component
	State          string
	Ready          bool
	RetryAt        time.Time
	CompletedAt    time.Time
}

type Command struct {
	OperationID    string
	Component      Component
	Action         string
	ReferenceEpoch uint64
	IdempotencyKey string
}

type Receipt struct {
	ID          string
	OperationID string
	Component   Component
	Action      string
	Ready       bool
	CompletedAt time.Time
}

type Blocker struct {
	Owner  string
	Code   string
	Digest string
}

type BlockerReceipt struct {
	ID             string
	OperationID    string
	ComponentID    string
	ReferenceEpoch uint64
	CatalogVersion string
	ZeroDigest     string
	Blockers       []Blocker
	CompletedAt    time.Time
}
