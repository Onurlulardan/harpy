package artifact

func validBinding(value BindingProof) bool {
	return value.Schema == BindingSchema && validUUID(value.EnvironmentID) &&
		validUUID(value.LicenseBindingID) && validDigest(value.FinalizationReceiptSHA256) &&
		validDigest(value.RecoveryTrustBundleSHA256) && validKID(value.ControlKID) &&
		validUUID(value.RegistryNonce) && validUUID(value.OperationID)
}
