package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
	"github.com/google/uuid"
)

var subscriptionFactKeys = []string{
	"boundary", "grace_ends_at", "period_end", "period_start",
	"state", "subscription_id", "trial_ends_at",
}

func decodeSubscriptionFact(document []byte) (subscriptionFact, error) {
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, 64<<10)
	if err != nil || !bytes.Equal(canonical, document) ||
		!exactJSONKeys(document, subscriptionFactKeys) {
		return subscriptionFact{}, fmt.Errorf("subscription fact is not canonical")
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	var value subscriptionFactDocument
	if err := decoder.Decode(&value); err != nil {
		return subscriptionFact{}, fmt.Errorf("decode subscription fact: %w", err)
	}
	subscriptionID, err := uuid.Parse(value.SubscriptionID)
	state := subscriptiondomain.State(value.State)
	if err != nil || subscriptionID == uuid.Nil || !state.Valid() {
		return subscriptionFact{}, fmt.Errorf("subscription fact identity is invalid")
	}
	period, err := decodeFactPeriod(value.PeriodStart, value.PeriodEnd)
	if err != nil {
		return subscriptionFact{}, err
	}
	trialEnd, err := decodeOptionalInstant(value.TrialEndsAt)
	if err != nil {
		return subscriptionFact{}, err
	}
	graceEnd, err := decodeOptionalInstant(value.GraceEndsAt)
	if err != nil {
		return subscriptionFact{}, err
	}
	fact := subscriptionFact{
		subscriptionID: subscriptionID, state: state, period: period,
		trialEndsAt: trialEnd, graceEndsAt: graceEnd,
	}
	if value.Boundary != nil {
		boundary, boundaryErr := decodeBoundaryFact(*value.Boundary, fact)
		if boundaryErr != nil {
			return subscriptionFact{}, boundaryErr
		}
		fact.boundary = &boundary
	}
	return fact, nil
}

func exactJSONKeys(document []byte, expected []string) bool {
	var fields map[string]json.RawMessage
	if json.Unmarshal(document, &fields) != nil || len(fields) != len(expected) {
		return false
	}
	for _, key := range expected {
		if _, exists := fields[key]; !exists {
			return false
		}
	}
	return true
}

func decodeFactPeriod(startValue, endValue *string) (*billingperiod.Period, error) {
	if startValue == nil && endValue == nil {
		return nil, nil
	}
	if startValue == nil || endValue == nil {
		return nil, fmt.Errorf("subscription fact period is incomplete")
	}
	start, err := decodeInstant(*startValue)
	if err != nil {
		return nil, err
	}
	end, err := decodeInstant(*endValue)
	if err != nil {
		return nil, err
	}
	period, err := billingperiod.New(start, end)
	if err != nil {
		return nil, fmt.Errorf("subscription fact period is invalid: %w", err)
	}
	return &period, nil
}

func decodeOptionalInstant(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := decodeInstant(*value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func decodeInstant(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || parsed.Location() != time.UTC ||
		parsed.Nanosecond()%int(time.Microsecond) != 0 ||
		parsed.Format(time.RFC3339Nano) != value {
		return time.Time{}, fmt.Errorf("subscription fact instant is invalid")
	}
	return parsed, nil
}
