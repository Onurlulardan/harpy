package catalogowner

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/analytics/api/catalogpayload"
)

const runtimePlanContract = "analytics.catalog-runtime-plan.v1"

type runtimePlan struct {
	Contract       string          `json:"contract"`
	Kind           string          `json:"kind"`
	Status         string          `json:"status"`
	ComponentID    string          `json:"component_id"`
	SourceContract string          `json:"source_contract"`
	DocumentSHA256 string          `json:"document_sha256"`
	BlockerCode    string          `json:"blocker_code,omitempty"`
	Sources        []string        `json:"sources"`
	Query          json.RawMessage `json:"query,omitempty"`
	Parameters     json.RawMessage `json:"parameters,omitempty"`
	Output         json.RawMessage `json:"output,omitempty"`
	Dashboard      json.RawMessage `json:"dashboard,omitempty"`
}

func buildRuntimePlan(
	document []byte, component contract.Component,
) ([]byte, [sha256.Size]byte, error) {
	verified, err := catalogpayload.Verify(
		document, component.ID, component.ContentSHA256,
	)
	if err != nil {
		return nil, [sha256.Size]byte{}, err
	}
	payload := verified.Value()
	plan := runtimePlan{Contract: runtimePlanContract, Kind: payload.Kind}
	if payload.Kind == "dashboard" {
		plan.Status = "candidate"
		plan.Dashboard = append(json.RawMessage(nil), payload.Definition...)
		plan.Sources, err = dashboardSources(payload.Definition)
	} else {
		plan, err = endpointRuntimePlan(payload.Definition)
	}
	if err != nil {
		return nil, [sha256.Size]byte{}, err
	}
	plan.ComponentID = component.ID
	plan.SourceContract = payload.SourceContract
	plan.DocumentSHA256 = component.ContentSHA256
	encoded, err := json.Marshal(plan)
	if err != nil {
		return nil, [sha256.Size]byte{}, fmt.Errorf("encode Analytics runtime plan: %w", err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(encoded, 4<<20)
	if err != nil {
		return nil, [sha256.Size]byte{}, fmt.Errorf("canonicalize Analytics runtime plan: %w", err)
	}
	return canonical, sha256.Sum256(canonical), nil
}

func dashboardSources(document []byte) ([]string, error) {
	var definition struct {
		DataSources []struct {
			EndpointSlug string `json:"endpointSlug"`
		} `json:"dataSources"`
	}
	if err := json.Unmarshal(document, &definition); err != nil ||
		len(definition.DataSources) == 0 {
		return nil, fmt.Errorf("Analytics dashboard sources are invalid")
	}
	result := make([]string, len(definition.DataSources))
	for index, source := range definition.DataSources {
		if source.EndpointSlug == "" {
			return nil, fmt.Errorf("Analytics dashboard source is invalid")
		}
		result[index] = source.EndpointSlug
	}
	return result, nil
}

func unqualified(value, alias string) (string, bool) {
	prefix := alias + "."
	if strings.HasPrefix(value, prefix) && len(value) > len(prefix) {
		return strings.TrimPrefix(value, prefix), true
	}
	if !strings.Contains(value, ".") && value != "" {
		return value, true
	}
	return "", false
}
