package v1authority

import (
	_ "embed"
	"fmt"
	"strings"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

//go:embed sql/platform/cutover_apply.sql
var cutoverApplySQL string

type CutoverContract struct {
	query string
}

func NewCutoverContract() (CutoverContract, error) {
	query := strings.TrimSpace(cutoverApplySQL)
	if query == "" {
		return CutoverContract{}, fmt.Errorf("IAM V1 authority cutover query is empty")
	}
	return CutoverContract{query: query}, nil
}

func (CutoverContract) ID() string {
	return cutoverv1.IAMV1ProtectedAuthorityAdoptionContractID
}

func (contract CutoverContract) Query() string {
	return contract.query
}
