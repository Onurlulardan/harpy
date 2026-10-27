package cutover

import (
	"fmt"
	"regexp"
	"strings"
)

var applyContractIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_.-]{0,127}$`)

type ApplyContract struct {
	id    string
	query string
}

func NewApplyContract(id string, query string) (ApplyContract, error) {
	query = strings.TrimSpace(query)
	if !applyContractIDPattern.MatchString(id) || query == "" {
		return ApplyContract{}, fmt.Errorf("cutover apply contract is invalid")
	}
	return ApplyContract{id: id, query: query}, nil
}

func (contract ApplyContract) ID() string {
	return contract.id
}

func (contract ApplyContract) Query() string {
	return contract.query
}
