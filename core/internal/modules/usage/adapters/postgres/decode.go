package postgres

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

func decodeDocument(document []byte, target any) error {
	if len(document) == 0 || len(document) > 4*1024*1024 {
		return domain.ErrUnavailable
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if decoder.Decode(target) != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return domain.ErrUnavailable
	}
	return nil
}

func validCount(values ...int64) bool {
	for _, value := range values {
		if value < 0 {
			return false
		}
	}
	return true
}

func validTime(value time.Time) bool {
	return !value.IsZero() && value.Nanosecond()%1000 == 0
}

func validMetric(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0
}

func decodeActiveUsers(document []byte) (domain.ActiveUsers, error) {
	var wire activeDocument
	if decodeDocument(document, &wire) != nil || !validCount(wire.DAU, wire.WAU, wire.MAU,
		wire.ConcurrentUsers, wire.TotalRequests) {
		return domain.ActiveUsers{}, domain.ErrUnavailable
	}
	return domain.ActiveUsers{DAU: wire.DAU, WAU: wire.WAU, MAU: wire.MAU,
		ConcurrentUsers: wire.ConcurrentUsers, TotalRequests: wire.TotalRequests}, nil
}

func decodeTopUsers(document []byte, scope domain.Scope) ([]domain.TopUser, error) {
	var wire []topUserDocument
	if decodeDocument(document, &wire) != nil || len(wire) > 50 {
		return nil, domain.ErrUnavailable
	}
	result := make([]domain.TopUser, 0, len(wire))
	for _, item := range wire {
		if _, err := domain.NewTimelineQuery(scope, item.SubjectID, 1, 1); err != nil ||
			!validCount(item.RequestCount, item.UniqueRoutes) || !validTime(item.LastActive) {
			return nil, domain.ErrUnavailable
		}
		result = append(result, domain.TopUser{SubjectID: item.SubjectID,
			RequestCount: item.RequestCount, UniqueRoutes: item.UniqueRoutes,
			LastActive: item.LastActive.UTC()})
	}
	return result, nil
}

func timelineLevel(status int) string {
	if status >= 500 {
		return "error"
	}
	if status >= 400 {
		return "warn"
	}
	return "info"
}

func validFactID(value string) bool {
	id, err := uuid.Parse(value)
	return err == nil && id.String() == value && id.Version() == 7
}
