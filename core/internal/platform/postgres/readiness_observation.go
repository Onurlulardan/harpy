package postgres

type ReadinessObservation struct {
	database       bool
	databaseRole   bool
	schemaContract bool
	schemaReason   SchemaReadinessReason
}

func (observation ReadinessObservation) DatabaseReady() bool {
	return observation.database
}

func (observation ReadinessObservation) DatabaseRoleReady() bool {
	return observation.databaseRole
}

func (observation ReadinessObservation) SchemaContractReady() bool {
	return observation.schemaContract
}

func (observation ReadinessObservation) SchemaReason() SchemaReadinessReason {
	return observation.schemaReason
}
