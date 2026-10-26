package buildinfo

import (
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func TestReadProfileBindsExactServiceIdentity(t *testing.T) {
	tests := []struct {
		name string
		read func(config.Environment) (Info, error)
		want runtimev1.Service
	}{
		{name: "api", read: ReadAPI, want: runtimev1.ServiceCoreAPI},
		{name: "worker", read: ReadWorker, want: runtimev1.ServiceCoreWorker},
		{name: "migrator", read: ReadMigrator, want: runtimev1.ServiceCoreMigrator},
		{name: "operator", read: ReadOperator, want: runtimev1.ServiceCoreOperator},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			info, err := test.read(config.EnvironmentDevelopment)
			if err != nil {
				t.Fatal(err)
			}
			if info.Service() != test.want {
				t.Fatalf("service = %q, want %q", info.Service(), test.want)
			}
			if info.Environment() != runtimev1.EnvironmentDevelopment {
				t.Fatalf("environment = %q", info.Environment())
			}
		})
	}
}
