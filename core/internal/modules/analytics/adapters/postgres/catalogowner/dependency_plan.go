package catalogowner

import (
	"encoding/json"
	"fmt"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
)

func (owner *Owner) FormsDependencySlugs(
	request contract.ProvisionRequest,
) ([]string, bool, error) {
	if owner == nil {
		return nil, false, fmt.Errorf(
			"Analytics Catalog owner is unavailable",
		)
	}
	planBytes, _, err := buildRuntimePlan(
		request.Document,
		request.Component,
	)
	if err != nil {
		return nil, false, err
	}
	var plan runtimePlan
	if err = json.Unmarshal(planBytes, &plan); err != nil {
		return nil, false, fmt.Errorf(
			"decode Analytics Catalog runtime plan: %w",
			err,
		)
	}
	if plan.Kind != "endpoint" {
		return nil, false, nil
	}
	return append([]string(nil), plan.Sources...), true, nil
}
