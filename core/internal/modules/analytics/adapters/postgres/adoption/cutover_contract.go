package adoption

import (
	_ "embed"
	"fmt"
	"strings"
)

const CutoverContractID = "analytics.v1_adoption.v1"

//go:embed sql/tenant/cutover_apply.sql
var cutoverApplySQL string

type CutoverContract struct {
	id    string
	query string
}

func NewCutoverContract() (CutoverContract, error) {
	query := strings.TrimSpace(cutoverApplySQL)
	if query == "" {
		return CutoverContract{}, fmt.Errorf("Analytics cutover contract query is empty")
	}
	return CutoverContract{id: CutoverContractID, query: query}, nil
}

func (contract CutoverContract) ID() string    { return contract.id }
func (contract CutoverContract) Query() string { return contract.query }
