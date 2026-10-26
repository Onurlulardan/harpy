package recordkernel

type CompileInput struct {
	Projection Projection
	Binding    CatalogBinding
	Registry   RegistryEvidence
	Budget     BudgetProfile
}
