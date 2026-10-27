package dashboardhttp

import "encoding/json"

type widgetDocument struct {
	WidgetID       string          `json:"widget_id"`
	QueryVersionID string          `json:"query_version_id"`
	Position       uint16          `json:"position"`
	Parameters     json.RawMessage `json:"parameters"`
	Output         json.RawMessage `json:"output"`
}

type createDocument struct {
	Name       string           `json:"name"`
	Visibility string           `json:"visibility"`
	Layout     json.RawMessage  `json:"layout"`
	Widgets    []widgetDocument `json:"widgets"`
}

type updateDocument struct {
	ExpectedRootRevision    uint64           `json:"expected_root_revision"`
	ExpectedVersionRevision uint64           `json:"expected_version_revision"`
	Name                    string           `json:"name"`
	Visibility              string           `json:"visibility"`
	Layout                  json.RawMessage  `json:"layout"`
	Widgets                 []widgetDocument `json:"widgets"`
}

type publishDocument struct {
	DraftVersionID        string `json:"draft_version_id"`
	ExpectedRootRevision  uint64 `json:"expected_root_revision"`
	ExpectedDraftRevision uint64 `json:"expected_draft_revision"`
}

type grantDocument struct {
	PrincipalID      string `json:"principal_id"`
	PrincipalKind    string `json:"principal_kind"`
	Level            string `json:"level"`
	ExpectedRevision uint64 `json:"expected_revision"`
}

type revokeDocument struct {
	PrincipalID      string `json:"principal_id"`
	PrincipalKind    string `json:"principal_kind"`
	ExpectedRevision uint64 `json:"expected_revision"`
}

type deleteDocument struct {
	ExpectedRootRevision uint64 `json:"expected_root_revision"`
}
