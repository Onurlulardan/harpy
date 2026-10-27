package postgresv1

import "fmt"

func validateOperationSteps(actual, expected []OperationStep) error {
	if actual == nil || len(actual) != len(expected) {
		return fmt.Errorf("operation step set is incomplete")
	}
	for index := range expected {
		if actual[index] != expected[index] {
			return fmt.Errorf("operation step %d differs from contract", index)
		}
	}
	return nil
}

func operationHeaderValid(version, expectedVersion, origin, database string) bool {
	return version == expectedVersion && DatabaseOrigin(origin).Valid() &&
		databasePattern.MatchString(database)
}

func step(id, class string) OperationStep {
	return OperationStep{StepID: id, TransactionClass: class}
}

func validateInputDigests(values ...string) error {
	for _, value := range values {
		if err := validateDigest(value, "operation input hash"); err != nil {
			return err
		}
	}
	return nil
}
