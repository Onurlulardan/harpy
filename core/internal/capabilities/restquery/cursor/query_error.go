package cursor

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func invalidCursorError() error {
	return cursorError(restqueryv1.ErrorQueryCursorInvalid)
}

func expiredCursorError() error {
	return cursorError(restqueryv1.ErrorQueryCursorExpired)
}

func staleCursorError() error {
	return cursorError(restqueryv1.ErrorQueryCursorStale)
}

func cursorError(code restqueryv1.ErrorCode) error {
	queryError, err := api.NewQueryError(code, nil)
	if err != nil {
		return errInvalidCodec
	}
	return queryError
}
