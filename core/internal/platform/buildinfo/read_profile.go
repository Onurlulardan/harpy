package buildinfo

import (
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func ReadAPI(environment config.Environment) (Info, error) {
	return readConfigured(runtimev1.ServiceCoreAPI, environment)
}

func ReadWorker(environment config.Environment) (Info, error) {
	return readConfigured(runtimev1.ServiceCoreWorker, environment)
}

func ReadMigrator(environment config.Environment) (Info, error) {
	return readConfigured(runtimev1.ServiceCoreMigrator, environment)
}

func ReadOperator(environment config.Environment) (Info, error) {
	return readConfigured(runtimev1.ServiceCoreOperator, environment)
}

func readConfigured(service runtimev1.Service, environment config.Environment) (Info, error) {
	return Read(service, runtimev1.Environment(environment.String()))
}
