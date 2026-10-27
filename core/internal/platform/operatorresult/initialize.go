package operatorresult

import (
	"io"

	operatorv1 "github.com/Onurlulardan/harpy/contracts/operator/v1"
	"github.com/google/uuid"
)

func WriteInitialize(
	writer io.Writer,
	noop bool,
	operationID [16]byte,
	tenantID [16]byte,
) error {
	status := operatorv1.InitializeCompleted
	if noop {
		status = operatorv1.InitializeNoop
	}
	wire, err := operatorv1.NewInitializeResult(
		status, uuid.UUID(operationID), uuid.UUID(tenantID),
	)
	if err != nil {
		return err
	}
	return operatorv1.WriteInitializeResult(writer, wire)
}
