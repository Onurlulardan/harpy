package executor

func sameStatementBinding(batch consumedBatch) bool {
	recordTenant, recordTenantErr := batch.record.TenantID()
	countTenant, countTenantErr := batch.exactCount.TenantID()
	recordActor, recordActorErr := batch.record.ActorID()
	countActor, countActorErr := batch.exactCount.ActorID()
	recordCatalog, recordCatalogErr := batch.record.Catalog()
	countCatalog, countCatalogErr := batch.exactCount.Catalog()
	recordAuthorization, recordAuthorizationErr :=
		batch.record.AuthorizationContextSHA256()
	countAuthorization, countAuthorizationErr :=
		batch.exactCount.AuthorizationContextSHA256()
	recordMandatory, recordMandatoryErr := batch.record.MandatorySHA256()
	countMandatory, countMandatoryErr := batch.exactCount.MandatorySHA256()
	recordIdentity, recordIdentityErr := batch.record.QueryIdentitySHA256()
	countIdentity, countIdentityErr := batch.exactCount.QueryIdentitySHA256()
	recordContinuation, recordContinuationErr := batch.record.ContinuationSHA256()
	countContinuation, countContinuationErr := batch.exactCount.ContinuationSHA256()
	return recordTenantErr == nil && countTenantErr == nil &&
		recordActorErr == nil && countActorErr == nil &&
		recordCatalogErr == nil && countCatalogErr == nil &&
		recordAuthorizationErr == nil && countAuthorizationErr == nil &&
		recordMandatoryErr == nil && countMandatoryErr == nil &&
		recordIdentityErr == nil && countIdentityErr == nil &&
		recordContinuationErr == nil && countContinuationErr == nil &&
		recordTenant == countTenant && recordActor == countActor &&
		recordCatalog == countCatalog && recordAuthorization == countAuthorization &&
		recordMandatory == countMandatory && recordIdentity == countIdentity &&
		recordContinuation == countContinuation
}
