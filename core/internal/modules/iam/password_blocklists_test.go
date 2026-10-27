package iam

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
)

func TestOpenPasswordBlocklistsBuildsBothFinalPolicySources(t *testing.T) {
	t.Setenv("HARPY_ENVIRONMENT", "test")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv("HARPY_POSTGRES_DSN", "postgres://user@example.invalid/db?sslmode=disable")
	t.Setenv("HARPY_REST_QUERY_CURSOR_KEYRING", passwordBlocklistCursorFixture)
	configfixture.SetAPI(t)
	settings, err := config.LoadAPI()
	if err != nil {
		t.Fatal(err)
	}
	iamSettings, _ := settings.IAM()
	configured, err := iamSettings.PasswordBlocklists()
	if err != nil {
		t.Fatal(err)
	}
	blocklists, err := OpenPasswordBlocklists(configured)
	if err != nil {
		t.Fatal(err)
	}
	defer blocklists.Close()
	common, err := blocklists.Common().Blocked(
		context.Background(), []byte("fixture common password"),
	)
	compromised, compromisedErr := blocklists.Compromised().Blocked(
		context.Background(), []byte("fixture compromised password"),
	)
	if err != nil || compromisedErr != nil || !common || !compromised {
		t.Fatalf("common=%t compromised=%t errors=%v/%v", common, compromised, err, compromisedErr)
	}
}

const passwordBlocklistCursorFixture = `{"version":1,"current":{"kid":"current-1","key":"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE"},"previous":{"kid":"previous-1","key":"AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI"}}`
