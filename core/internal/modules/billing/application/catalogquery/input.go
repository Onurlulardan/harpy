package catalogquery

import (
	"encoding/base64"
	"time"

	offeringdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/offering"
)

const MaximumPageSize = 100

type ListRequest struct {
	limit   int
	after   string
	instant time.Time
}

func (request ListRequest) Limit() int         { return request.limit }
func (request ListRequest) Instant() time.Time { return request.instant }

type GetRequest struct {
	key     string
	instant time.Time
}

func (request GetRequest) Key() string        { return request.key }
func (request GetRequest) Instant() time.Time { return request.instant }

type ActiveCampaignRequest struct{ instant time.Time }

func (request ActiveCampaignRequest) Instant() time.Time { return request.instant }

func encodeCursor(key string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(key))
}

func decodeCursor(cursor string) (string, error) {
	if cursor == "" {
		return "", nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil || base64.RawURLEncoding.EncodeToString(decoded) != cursor {
		return "", ErrRequest
	}
	key, err := offeringdomain.NewKey(string(decoded))
	if err != nil {
		return "", ErrRequest
	}
	return string(key), nil
}

func validInstant(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1_000 == 0
}
