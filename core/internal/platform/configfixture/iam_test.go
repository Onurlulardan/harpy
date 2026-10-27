package configfixture_test

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
)

const cursorKeyring = `{"version":1,"current":{"kid":"current-1","key":"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE"},"previous":{"kid":"previous-1","key":"AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI"}}`

func TestIAMAPIFixtureLoadsThroughActualProfile(t *testing.T) {
	setRuntimeEnvironment(t)
	t.Setenv("HARPY_REST_QUERY_CURSOR_KEYRING", cursorKeyring)
	configfixture.SetAPI(t)
	loaded, err := config.LoadAPI()
	if err != nil {
		t.Fatal(err)
	}
	iamConfig, err := loaded.IAM()
	if err != nil || iamConfig.Validate() != nil {
		t.Fatalf("fixture did not produce valid IAM API config: %v", err)
	}
}

func TestIAMWorkerFixtureLoadsThroughActualProfile(t *testing.T) {
	setRuntimeEnvironment(t)
	t.Setenv("HARPY_RABBITMQ_URL", "amqp://worker:secret@rabbit.invalid/%2f")
	configfixture.SetWorker(t)
	loaded, err := config.LoadWorker()
	if err != nil {
		t.Fatal(err)
	}
	iamConfig, err := loaded.IAM()
	if err != nil || iamConfig.Validate() != nil {
		t.Fatalf("fixture did not produce valid IAM worker config: %v", err)
	}
}

func setRuntimeEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("HARPY_ENVIRONMENT", "test")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv(
		"HARPY_POSTGRES_DSN",
		"postgresql://user@example.invalid/harpy?sslmode=disable",
	)
}
