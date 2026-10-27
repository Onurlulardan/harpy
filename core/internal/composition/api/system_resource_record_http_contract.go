package api

import (
	"context"
	"encoding/json"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

type systemResourceRecordHTTPAccessVerifier interface {
	Verify(context.Context, systemresources.LayoutAccessRequest) (
		systemresources.LayoutAccessDecision, error,
	)
}

type systemResourceRecordHTTPApplication interface {
	Fields(context.Context, systemresources.Admission, systemresources.Resource) (
		[]systemResourceRecordHTTPField, error,
	)
	PrepareCreate(context.Context, systemresources.CreateInput) (
		systemresources.PreparedCreate, error,
	)
	Create(context.Context, systemresources.Admission, systemresources.PreparedCreate) (
		systemResourceRecordHTTPView, error,
	)
	Get(context.Context, systemresources.Admission, systemresources.GetInput) (
		systemResourceRecordHTTPView, error,
	)
	Query(context.Context, systemresources.Admission, systemresources.QueryInput) (
		systemResourceRecordHTTPQueryView, error,
	)
	Update(context.Context, systemresources.Admission, systemresources.UpdateInput) (
		systemResourceRecordHTTPView, error,
	)
	Delete(context.Context, systemresources.Admission, systemresources.DeleteInput) error
}

type systemResourceRecordHTTPField struct {
	ID       systemresources.ID
	Name     string
	Editable bool
}

type systemResourceRecordHTTPView struct {
	Data     map[string]json.RawMessage
	ID       string
	Revision int64
}

type systemResourceRecordHTTPQueryView struct {
	Items      []json.RawMessage
	HasMore    bool
	NextCursor string
	ExactCount *int64
}

const (
	systemResourceRecordHTTPRead   = "read"
	systemResourceRecordHTTPCreate = "create"
	systemResourceRecordHTTPUpdate = "update"
	systemResourceRecordHTTPDelete = "delete"
)
