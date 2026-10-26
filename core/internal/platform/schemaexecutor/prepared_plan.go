package schemaexecutor

type PreparedPlan struct {
	canonical []byte
	sha256    string
	steps     []PreparedStep
}

type PreparedStep struct {
	identifierMapping []byte
	arguments         []byte
	canonicalPreimage []byte
	sha256            string
}

func (plan PreparedPlan) CanonicalJSON() []byte {
	return append([]byte(nil), plan.canonical...)
}

func (plan PreparedPlan) SHA256() string { return plan.sha256 }

func (plan PreparedPlan) Steps() []PreparedStep {
	return append([]PreparedStep(nil), plan.steps...)
}

func (step PreparedStep) IdentifierMappingJSON() []byte {
	return append([]byte(nil), step.identifierMapping...)
}

func (step PreparedStep) ArgumentsJSON() []byte {
	return append([]byte(nil), step.arguments...)
}

func (step PreparedStep) CanonicalPreimage() []byte {
	return append([]byte(nil), step.canonicalPreimage...)
}

func (step PreparedStep) SHA256() string { return step.sha256 }
