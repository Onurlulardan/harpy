package offering

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
)

const canonicalTimestamp = "2006-01-02T15:04:05.000000Z"

func publicFilter(snapshotAt time.Time) (restauthorization.MandatoryFilter, error) {
	active, err := enumPredicate(fieldLifecycle, api.OperatorEq, "active")
	if err != nil {
		return restauthorization.MandatoryFilter{}, ErrContract
	}
	published, err := enumPredicate(
		fieldPublicationState, api.OperatorEq, "published",
	)
	if err != nil {
		return restauthorization.MandatoryFilter{}, ErrContract
	}
	visible, err := booleanPredicate(fieldPublic, api.OperatorEq, true)
	if err != nil {
		return restauthorization.MandatoryFilter{}, ErrContract
	}
	available, err := booleanPredicate(fieldAvailable, api.OperatorEq, true)
	if err != nil {
		return restauthorization.MandatoryFilter{}, ErrContract
	}
	started, err := timestampPredicate(
		fieldEffectiveStart, api.OperatorLTE, snapshotAt,
	)
	if err != nil {
		return restauthorization.MandatoryFilter{}, ErrContract
	}
	ending, err := effectiveEndPredicate(snapshotAt)
	if err != nil {
		return restauthorization.MandatoryFilter{}, ErrContract
	}
	filter, err := restauthorization.NewMandatoryAnd([]restauthorization.MandatoryFilter{
		active, published, visible, available, started, ending,
	})
	if err != nil {
		return restauthorization.MandatoryFilter{}, ErrContract
	}
	return filter, nil
}

func effectiveEndPredicate(
	snapshotAt time.Time,
) (restauthorization.MandatoryFilter, error) {
	missing, err := predicateFilter(
		fieldEffectiveEnd, api.LogicalTypeTimestamp, api.OperatorIsNull,
		[]restauthorization.Value{},
	)
	if err != nil {
		return restauthorization.MandatoryFilter{}, ErrContract
	}
	future, err := timestampPredicate(fieldEffectiveEnd, api.OperatorGT, snapshotAt)
	if err != nil {
		return restauthorization.MandatoryFilter{}, ErrContract
	}
	filter, err := restauthorization.NewMandatoryOr(
		[]restauthorization.MandatoryFilter{missing, future},
	)
	if err != nil {
		return restauthorization.MandatoryFilter{}, ErrContract
	}
	return filter, nil
}
