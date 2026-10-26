package idempotency

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type StoredReceipt struct {
	ReceiptID            uuid.UUID
	RequestSHA256        string
	AuthorizationVersion int64
	RetentionClass       RetentionClass
	Status               int
	HeadersJSON          []byte
	Body                 []byte
	ResultReference      *string
	ResponseSHA256       string
	AsyncJobID           *uuid.UUID
	CompletedAt          time.Time
	RetainUntil          time.Time
}

type Receipt struct {
	id                   uuid.UUID
	requestSHA256        string
	authorizationVersion int64
	retentionClass       RetentionClass
	response             Response
	asyncJobID           *uuid.UUID
	completedAt          time.Time
	retainUntil          time.Time
}

func RestoreReceipt(stored StoredReceipt) (Receipt, error) {
	var headers map[string]string
	if err := json.Unmarshal(stored.HeadersJSON, &headers); err != nil {
		return Receipt{}, ErrInvalidReceipt
	}
	var response Response
	var err error
	if stored.Body != nil {
		response, err = NewBodyResponse(stored.Status, headers, stored.Body)
	} else if stored.ResultReference != nil {
		response, err = NewReferenceResponse(stored.Status, headers, *stored.ResultReference)
	} else {
		return Receipt{}, ErrInvalidReceipt
	}
	if err != nil || response.SHA256() != stored.ResponseSHA256 ||
		stored.ReceiptID == uuid.Nil || !sha256Pattern.MatchString(stored.RequestSHA256) ||
		stored.AuthorizationVersion < 1 || !stored.RetentionClass.valid() ||
		stored.CompletedAt.IsZero() || !stored.RetainUntil.After(stored.CompletedAt) ||
		(stored.RetentionClass == AsyncHandle) != (stored.AsyncJobID != nil) {
		return Receipt{}, ErrInvalidReceipt
	}
	return Receipt{
		id: stored.ReceiptID, requestSHA256: stored.RequestSHA256,
		authorizationVersion: stored.AuthorizationVersion,
		retentionClass:       stored.RetentionClass, response: response,
		asyncJobID:  cloneUUID(stored.AsyncJobID),
		completedAt: stored.CompletedAt.UTC(), retainUntil: stored.RetainUntil.UTC(),
	}, nil
}

func cloneUUID(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func (receipt Receipt) ID() uuid.UUID                  { return receipt.id }
func (receipt Receipt) RequestSHA256() string          { return receipt.requestSHA256 }
func (receipt Receipt) AuthorizationVersion() int64    { return receipt.authorizationVersion }
func (receipt Receipt) RetentionClass() RetentionClass { return receipt.retentionClass }
func (receipt Receipt) Response() Response             { return receipt.response }
func (receipt Receipt) AsyncJobID() *uuid.UUID         { return cloneUUID(receipt.asyncJobID) }
func (receipt Receipt) CompletedAt() time.Time         { return receipt.completedAt }
func (receipt Receipt) RetainUntil() time.Time         { return receipt.retainUntil }
