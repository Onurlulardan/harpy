package billingv1

import (
	"encoding/json"
	"testing"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestCampaignEndDateDistinguishesAbsentNullAndValueOnUpdate(t *testing.T) {
	oldEnd := "2026-08-01T00:00:00.000000Z"
	current := billingapi.CampaignView{
		ID: "campaign", Key: "summer", Revision: 2, Name: "Summer",
		Available: true, StartsAt: "2026-07-01T00:00:00.000000Z", EndsAt: &oldEnd,
	}
	tests := []struct {
		name     string
		document string
		want     *string
	}{
		{name: "absent", document: `{}`, want: &oldEnd},
		{name: "explicit null", document: `{"end_date":null}`},
		{name: "replacement", document: `{"end_date":"2026-09-02T03:04:05Z"}`,
			want: stringPointer("2026-09-02T03:04:05.000000Z")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := decodeCampaignMappingRequest(t, test.document)
			command, err := mapCampaignUpdate(
				billingapi.Access{}, "key", current, input, time.Now(),
			)
			if err != nil {
				t.Fatal(err)
			}
			assertOptionalString(t, "end_date", command.Version.EndsAt, test.want)
		})
	}
}

func TestCampaignEndDateMapsCreateNullAndAbsentToNil(t *testing.T) {
	for _, document := range []string{
		`{"name":"Summer","slug":"summer","start_date":"2026-07-01T00:00:00Z"}`,
		`{"name":"Summer","slug":"summer","start_date":"2026-07-01T00:00:00Z","end_date":null}`,
	} {
		input := decodeCampaignMappingRequest(t, document)
		command, err := mapCampaignCreate(billingapi.Access{}, "key", input)
		if err != nil {
			t.Fatal(err)
		}
		if command.InitialVersion.EndsAt != nil {
			t.Fatalf("document=%s end=%v", document, command.InitialVersion.EndsAt)
		}
	}
}

func decodeCampaignMappingRequest(t *testing.T, document string) campaignRequest {
	t.Helper()
	var input campaignRequest
	if err := json.Unmarshal([]byte(document), &input); err != nil {
		t.Fatal(err)
	}
	return input
}
