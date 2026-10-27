package schemaexecutor

import (
	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func buildPreparedStep(
	contractStep schemav1.PreparedStep,
	step schemav1.Step,
) (PreparedStep, error) {
	identifiers, err := marshalCanonical(step.IdentifierMapping)
	if err != nil {
		return PreparedStep{}, err
	}
	arguments, err := marshalCanonical(step.Arguments)
	if err != nil {
		return PreparedStep{}, err
	}
	return PreparedStep{
		identifierMapping: identifiers,
		arguments:         arguments,
		canonicalPreimage: contractStep.CanonicalPreimage(),
		sha256:            contractStep.SHA256(),
	}, nil
}

func marshalCanonical(value any) ([]byte, error) {
	return attestation.MarshalCanonicalSchemaPlan(value)
}
