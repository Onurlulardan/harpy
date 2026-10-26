package cutoverv1

func AuthorityDependencyContractSHA256(
	dependencies []AuthorityCatalogDependency,
) string {
	return authorityCatalogDigest(dependencies)
}

func AuthorityReceiptDependencyContractSHA256(
	dependencies []AuthorityCatalogDependencyReceipt,
) string {
	contract := make([]AuthorityCatalogDependency, len(dependencies))
	for index, dependency := range dependencies {
		contract[index] = dependency.AuthorityCatalogDependency
	}
	return AuthorityDependencyContractSHA256(contract)
}

func AuthorityReceiptDependencyInstanceSHA256(
	dependencies []AuthorityCatalogDependencyReceipt,
) string {
	return authorityCatalogDigest(dependencies)
}
