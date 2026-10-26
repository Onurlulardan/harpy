package executor

import (
	"crypto/sha256"
	"fmt"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/aggregatepage"
)

type AggregateResult struct {
	page      aggregatepage.Page
	stateSeal [sha256.Size]byte
	valid     bool
}

func newAggregateResult(page aggregatepage.Page) (AggregateResult, error) {
	result := AggregateResult{page: page, valid: true}
	seal, err := aggregateResultSeal(result)
	if err != nil {
		return AggregateResult{}, ErrContract
	}
	result.stateSeal = seal
	if result.Validate() != nil {
		return AggregateResult{}, ErrContract
	}
	return result, nil
}

func (result AggregateResult) Validate() error {
	if !result.valid || result.page.Validate() != nil ||
		result.stateSeal == [sha256.Size]byte{} {
		return ErrContract
	}
	seal, err := aggregateResultSeal(result)
	if err != nil || seal != result.stateSeal {
		return ErrContract
	}
	return nil
}

func (result AggregateResult) Groups() ([][]byte, error) {
	if result.Validate() != nil {
		return nil, ErrContract
	}
	groups, err := result.page.Groups()
	if err != nil {
		return nil, ErrContract
	}
	return groups, nil
}

func (result AggregateResult) Limit() (int, error) {
	if result.Validate() != nil {
		return 0, ErrContract
	}
	limit, err := result.page.Limit()
	if err != nil {
		return 0, ErrContract
	}
	return limit, nil
}

func (AggregateResult) String() string   { return redacted }
func (AggregateResult) GoString() string { return redacted }
func (AggregateResult) LogValue() slog.Value {
	return slog.StringValue(redacted)
}
func (result AggregateResult) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (AggregateResult) MarshalJSON() ([]byte, error)          { return nil, ErrDirectMarshal }
