package executor

import (
	"crypto/sha256"
	"fmt"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/recordpage"
)

type Result struct {
	page       recordpage.Page
	exactCount int64
	stateSeal  [sha256.Size]byte
	hasCount   bool
	valid      bool
}

func newResult(
	page recordpage.Page,
	exactCount int64,
	hasCount bool,
) (Result, error) {
	result := Result{
		page: page, exactCount: exactCount, hasCount: hasCount, valid: true,
	}
	seal, err := resultSeal(result)
	if err != nil {
		return Result{}, ErrContract
	}
	result.stateSeal = seal
	if result.Validate() != nil {
		return Result{}, ErrContract
	}
	return result, nil
}

func (result Result) Validate() error {
	if !result.valid || result.page.Validate() != nil || result.exactCount < 0 ||
		!result.hasCount && result.exactCount != 0 ||
		result.stateSeal == [sha256.Size]byte{} {
		return ErrContract
	}
	seal, err := resultSeal(result)
	if err != nil || seal != result.stateSeal {
		return ErrContract
	}
	return nil
}

func (result Result) Page() (recordpage.Page, error) {
	if result.Validate() != nil {
		return recordpage.Page{}, ErrContract
	}
	return result.page, nil
}

func (result Result) ExactCount() (int64, bool, error) {
	if result.Validate() != nil {
		return 0, false, ErrContract
	}
	return result.exactCount, result.hasCount, nil
}

func (Result) String() string   { return redacted }
func (Result) GoString() string { return redacted }
func (Result) LogValue() slog.Value {
	return slog.StringValue(redacted)
}
func (result Result) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (Result) MarshalJSON() ([]byte, error)          { return nil, ErrDirectMarshal }
