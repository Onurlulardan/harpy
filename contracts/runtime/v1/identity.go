package runtimev1

const (
	ContractIdentity = "platform-runtime.v1"
	ContractVersion  = 1
	GoVersion        = "go1.26.5"
)

type Service string

const (
	ServiceCoreAPI      Service = "core-api"
	ServiceCoreMigrator Service = "core-migrator"
	ServiceCoreOperator Service = "core-operator"
	ServiceCoreWorker   Service = "core-worker"
)

func (service Service) Valid() bool {
	switch service {
	case ServiceCoreAPI, ServiceCoreMigrator, ServiceCoreOperator, ServiceCoreWorker:
		return true
	default:
		return false
	}
}

type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentProduction  Environment = "production"
	EnvironmentStaging     Environment = "staging"
	EnvironmentTest        Environment = "test"
)

func (environment Environment) Valid() bool {
	switch environment {
	case EnvironmentDevelopment, EnvironmentProduction,
		EnvironmentStaging, EnvironmentTest:
		return true
	default:
		return false
	}
}
