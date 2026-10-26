package httpquery

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/internal/jsonscan"
)

const maximumJSONMemberNesting = jsonscan.MaximumNesting

var (
	errInvalidJSONDocument = errors.New("JSON member scan failed")
	errJSONNestingLimit    = errors.New("JSON member scan exceeded nesting limit")
)

type memberScan struct {
	duplicateMemberPaths []string
}

func (s memberScan) duplicatePaths() []string {
	return append([]string(nil), s.duplicateMemberPaths...)
}

func scanJSONMembers(document []byte) (memberScan, error) {
	result, err := jsonscan.Scan(document, maximumJSONMemberNesting)
	if err != nil {
		if errors.Is(err, jsonscan.ErrNestingLimit) {
			return memberScan{}, errJSONNestingLimit
		}
		return memberScan{}, errInvalidJSONDocument
	}
	return memberScan{duplicateMemberPaths: result.DuplicatePointers()}, nil
}
