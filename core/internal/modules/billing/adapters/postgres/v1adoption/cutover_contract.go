package v1adoption

import (
	_ "embed"
	"fmt"
	"strings"
)

const (
	GlobalCutoverContractID = "billing.v1_adoption.global.v1"
	TenantCutoverContractID = "billing.v1_adoption.tenant.v1"
	BaseCutoverContractID   = "billing.v1_base_adoption.v1"
)

//go:embed sql/platform/cutover_global_apply.sql
var globalCutoverApplySQL string

//go:embed sql/platform/cutover_tenant_apply.sql
var tenantCutoverApplySQL string

//go:embed sql/platform/cutover_base_apply.sql
var baseCutoverApplySQL string

type CutoverContract struct {
	id    string
	query string
}

func NewGlobalCutoverContract() (CutoverContract, error) {
	return newCutoverContract(GlobalCutoverContractID, globalCutoverApplySQL)
}

func NewTenantCutoverContract() (CutoverContract, error) {
	return newCutoverContract(TenantCutoverContractID, tenantCutoverApplySQL)
}

func NewBaseCutoverContract() (CutoverContract, error) {
	return newCutoverContract(BaseCutoverContractID, baseCutoverApplySQL)
}

func newCutoverContract(id, query string) (CutoverContract, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return CutoverContract{}, fmt.Errorf("Billing cutover contract query is empty")
	}
	return CutoverContract{id: id, query: query}, nil
}

func (contract CutoverContract) ID() string    { return contract.id }
func (contract CutoverContract) Query() string { return contract.query }
