package schemaexecutor

import (
	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

type ValidatedStep struct {
	contractVersion string
	value           schemav1.Step
	v2Value         schemav2.Step
	arguments       any
	hash            string
}

func (step ValidatedStep) Operation() schemav1.OperationID {
	if step.contractVersion == schemav2.PlanContractVersion {
		return schemav1.OperationID(step.v2Value.OperationID)
	}
	return step.value.OperationID
}

func (step ValidatedStep) TransactionClass() schemav1.TransactionClass {
	if step.contractVersion == schemav2.PlanContractVersion {
		return schemav1.TransactionClass(step.v2Value.TransactionClass)
	}
	return step.value.TransactionClass
}

func (step ValidatedStep) ExecutorKind() schemav1.ExecutorKind {
	if step.contractVersion == schemav2.PlanContractVersion {
		return schemav1.ExecutorKind(step.v2Value.ExecutorKind)
	}
	return step.value.ExecutorKind
}

func (step ValidatedStep) Mapping() schemav1.IdentifierMapping {
	return step.value.IdentifierMapping
}

func (step ValidatedStep) V2Mapping() (schemav2.IdentifierMapping, bool) {
	if step.contractVersion != schemav2.PlanContractVersion {
		return schemav2.IdentifierMapping{}, false
	}
	return step.v2Value.IdentifierMapping, true
}

func (step ValidatedStep) ContractVersion() string { return step.contractVersion }

func (step ValidatedStep) ArgumentValue() any { return step.arguments }

func (step ValidatedStep) Ordinal() int {
	if step.contractVersion == schemav2.PlanContractVersion {
		return step.v2Value.Ordinal
	}
	return step.value.Ordinal
}

func (step ValidatedStep) HashSHA256() string { return step.hash }

type ValidatedPlan struct {
	steps   []ValidatedStep
	current int
}

func (plan ValidatedPlan) CurrentStep() ValidatedStep {
	return plan.steps[plan.current]
}

func (plan ValidatedPlan) StepCount() int { return len(plan.steps) }

func (plan ValidatedPlan) StepAt(ordinal int) (ValidatedStep, bool) {
	if ordinal < 0 || ordinal >= len(plan.steps) {
		return ValidatedStep{}, false
	}
	return plan.steps[ordinal], true
}
