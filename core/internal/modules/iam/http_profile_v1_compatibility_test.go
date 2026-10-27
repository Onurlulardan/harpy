package iam

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestProfileHTTPKeepsV1GetAndPutSuccessEnvelope(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPut} {
		t.Run(method, func(t *testing.T) {
			fixture := newProfileHTTPTestFixture(t)
			operation, body := "GetSelf", ""
			if method == http.MethodPut {
				operation, body = "UpdateSelf", `{"first_name":"Ada"}`
			}
			fixture.app.results[operation] = profileHTTPTestSelfView()
			response := fixture.serve(fixture.request(method, "/api/profile", body))
			if response.Code != http.StatusOK {
				t.Fatalf("status=%d body=%s", response.Code, response.Body)
			}
			var envelope struct {
				Success bool            `json:"success"`
				Data    json.RawMessage `json:"data"`
			}
			if json.Unmarshal(response.Body.Bytes(), &envelope) != nil ||
				!envelope.Success || !json.Valid(envelope.Data) {
				t.Fatalf("V1 envelope=%s", response.Body)
			}
			var data map[string]json.RawMessage
			if json.Unmarshal(envelope.Data, &data) != nil ||
				len(data["id"]) == 0 || len(data["email"]) == 0 {
				t.Fatalf("V1 profile data=%s", envelope.Data)
			}
		})
	}
}
