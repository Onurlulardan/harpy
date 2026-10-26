package lifecycle

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

type OwnerReceipt struct {
	ReceiptID       packagecontract.StableID
	OperationID     packagecontract.StableID
	ComponentID     packagecontract.ComponentID
	Owner           string
	OwnerGeneration int64
	Contract        string
	ContentSHA256   string
	Action          string
	Ready           bool
	CompletedAt     time.Time
}

func (value OwnerReceipt) Matches(operationID packagecontract.StableID, pin ComponentPin, action string) bool {
	return value.ReceiptID.Valid() && operationID.Valid() &&
		value.OperationID == operationID && value.ComponentID.String() == pin.ComponentID.String() &&
		value.Owner == pin.Owner && value.OwnerGeneration == pin.OwnerGeneration &&
		value.Contract == pin.Contract && value.ContentSHA256 == pin.ContentSHA256 &&
		value.Action == action && value.Ready && canonicalTime(value.CompletedAt)
}
