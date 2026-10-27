package domain

import "errors"

var (
	ErrUploadInvalid    = errors.New("documents upload input is invalid")
	ErrUploadTransition = errors.New("documents upload transition is invalid")
	ErrUploadConflict   = errors.New("documents current version conflicts")
)

type UploadState string

const (
	UploadCreated    UploadState = "created"
	UploadReceiving  UploadState = "receiving"
	UploadRetryable  UploadState = "retryable"
	UploadScanning   UploadState = "scanning"
	UploadReady      UploadState = "ready"
	UploadRejected   UploadState = "rejected"
	UploadConflicted UploadState = "conflicted"
	UploadCancelled  UploadState = "cancelled"
	UploadExpired    UploadState = "expired"
)

type Upload struct {
	ID                       ID
	TenantID                 ID
	ActorID                  ID
	DocumentID               ID
	CandidateVersionID       ID
	ExpectedCurrentVersionID ID
	RequestSHA256            [32]byte
	ExpectedContentSHA256    [32]byte
	ExpectedBytes            int64
	AttemptGeneration        uint64
	State                    UploadState
	CleanupRequired          bool
	CleanupReceiptSHA256     [32]byte
}

type UploadInput struct {
	ID, TenantID, ActorID, DocumentID, CandidateVersionID ID
	ExpectedCurrentVersionID                              ID
	RequestSHA256, ExpectedContentSHA256                  [32]byte
	ExpectedBytes                                         int64
}

func NewUpload(input UploadInput) (Upload, error) {
	if !input.ID.UUIDv7() || !input.TenantID.Valid() || !input.ActorID.Valid() ||
		!input.DocumentID.UUIDv7() || !input.CandidateVersionID.UUIDv7() ||
		input.RequestSHA256 == [32]byte{} || input.ExpectedContentSHA256 == [32]byte{} ||
		input.ExpectedBytes < 1 || input.ExpectedBytes > MaxUploadBytes ||
		input.ExpectedCurrentVersionID.Valid() && !input.ExpectedCurrentVersionID.UUIDv7() {
		return Upload{}, ErrUploadInvalid
	}
	return Upload{ID: input.ID, TenantID: input.TenantID, ActorID: input.ActorID,
		DocumentID: input.DocumentID, CandidateVersionID: input.CandidateVersionID,
		ExpectedCurrentVersionID: input.ExpectedCurrentVersionID,
		RequestSHA256:            input.RequestSHA256, ExpectedContentSHA256: input.ExpectedContentSHA256,
		ExpectedBytes: input.ExpectedBytes, State: UploadCreated}, nil
}

func (upload Upload) BeginAttempt() (Upload, error) {
	if upload.State != UploadCreated && upload.State != UploadRetryable || upload.CleanupRequired {
		return Upload{}, ErrUploadTransition
	}
	upload.AttemptGeneration++
	upload.State = UploadReceiving
	return upload, nil
}

func (upload Upload) CompleteStream(bytes int64, digest [32]byte) (Upload, error) {
	if upload.State != UploadReceiving {
		return Upload{}, ErrUploadTransition
	}
	if bytes != upload.ExpectedBytes || digest != upload.ExpectedContentSHA256 {
		upload.State = UploadRejected
		upload.CleanupRequired = true
		return upload, ErrUploadInvalid
	}
	upload.State = UploadScanning
	return upload, nil
}

func (upload Upload) RetryAfterTransient() (Upload, error) {
	if upload.State != UploadReceiving && upload.State != UploadScanning {
		return Upload{}, ErrUploadTransition
	}
	upload.State = UploadRetryable
	upload.CleanupRequired = true
	return upload, nil
}

func (upload Upload) ConfirmCleanup(receipt [32]byte) (Upload, error) {
	if !upload.CleanupRequired || receipt == [32]byte{} ||
		upload.State != UploadRetryable && upload.State != UploadRejected &&
			upload.State != UploadConflicted {
		return Upload{}, ErrUploadTransition
	}
	upload.CleanupRequired = false
	upload.CleanupReceiptSHA256 = receipt
	return upload, nil
}

func (upload Upload) AcceptReady(currentVersion ID) (Upload, error) {
	if upload.State == UploadReady {
		return upload, nil
	}
	if upload.State != UploadScanning {
		return Upload{}, ErrUploadTransition
	}
	if currentVersion != upload.ExpectedCurrentVersionID {
		upload.State = UploadConflicted
		upload.CleanupRequired = true
		return upload, ErrUploadConflict
	}
	upload.State = UploadReady
	return upload, nil
}

func (upload Upload) RejectContent() (Upload, error) {
	if upload.State != UploadScanning {
		return Upload{}, ErrUploadTransition
	}
	upload.State = UploadRejected
	upload.CleanupRequired = true
	return upload, nil
}

func (upload Upload) Cancel() (Upload, error) {
	if upload.State != UploadCreated && upload.State != UploadRetryable || upload.CleanupRequired {
		return Upload{}, ErrUploadTransition
	}
	upload.State = UploadCancelled
	return upload, nil
}
