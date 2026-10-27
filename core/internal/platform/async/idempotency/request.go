package idempotency

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

var sha256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type Request struct {
	key                  string
	requestSHA256        string
	authorizationVersion int64
	retentionClass       RetentionClass
	asyncJobID           *uuid.UUID
	requestedRetainUntil *time.Time
}

func NewWriteRequest(
	key string,
	requestSHA256 string,
	authorizationVersion int64,
	class RetentionClass,
) (Request, error) {
	if class != InteractiveWrite && class != AdministrativeWrite {
		return Request{}, ErrInvalidRequest
	}
	return newRequest(key, requestSHA256, authorizationVersion, class, nil, nil)
}

// NewAsyncHandleRequest accepts the caller-owned artifact horizon, including
// its required 24-hour tail, as retain-until. PostgreSQL also applies 30 days.
func NewAsyncHandleRequest(
	key string,
	requestSHA256 string,
	authorizationVersion int64,
	jobID uuid.UUID,
	requestedRetainUntil time.Time,
) (Request, error) {
	if jobID == uuid.Nil || requestedRetainUntil.IsZero() {
		return Request{}, ErrInvalidRequest
	}
	value := requestedRetainUntil.UTC()
	return newRequest(
		key, requestSHA256, authorizationVersion, AsyncHandle, &jobID, &value,
	)
}

func newRequest(
	key string,
	requestSHA256 string,
	authorizationVersion int64,
	class RetentionClass,
	jobID *uuid.UUID,
	retainUntil *time.Time,
) (Request, error) {
	if len(key) < 1 || len([]byte(key)) > 512 ||
		!sha256Pattern.MatchString(requestSHA256) || authorizationVersion < 1 || !class.valid() {
		return Request{}, ErrInvalidRequest
	}
	return Request{
		key: key, requestSHA256: requestSHA256,
		authorizationVersion: authorizationVersion, retentionClass: class,
		asyncJobID: jobID, requestedRetainUntil: retainUntil,
	}, nil
}

func (request Request) Validate() error {
	if len(request.key) < 1 || len([]byte(request.key)) > 512 ||
		!sha256Pattern.MatchString(request.requestSHA256) ||
		request.authorizationVersion < 1 || !request.retentionClass.valid() {
		return ErrInvalidRequest
	}
	if request.retentionClass == AsyncHandle {
		if request.asyncJobID == nil || *request.asyncJobID == uuid.Nil ||
			request.requestedRetainUntil == nil || request.requestedRetainUntil.IsZero() {
			return ErrInvalidRequest
		}
		return nil
	}
	if request.asyncJobID != nil || request.requestedRetainUntil != nil {
		return ErrInvalidRequest
	}
	return nil
}

func (request Request) Key() string                    { return request.key }
func (request Request) RequestSHA256() string          { return request.requestSHA256 }
func (request Request) AuthorizationVersion() int64    { return request.authorizationVersion }
func (request Request) RetentionClass() RetentionClass { return request.retentionClass }

func (request Request) AsyncJobID() *uuid.UUID {
	if request.asyncJobID == nil {
		return nil
	}
	value := *request.asyncJobID
	return &value
}

func (request Request) RequestedRetainUntil() *time.Time {
	if request.requestedRetainUntil == nil {
		return nil
	}
	value := *request.requestedRetainUntil
	return &value
}
