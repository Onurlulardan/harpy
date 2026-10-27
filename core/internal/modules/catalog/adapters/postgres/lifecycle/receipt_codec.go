package lifecycle

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

func receiptDigest(value catalogdomain.OwnerReceipt) string {
	document := strings.Join([]string{value.ReceiptID.String(), value.OperationID.String(),
		value.ComponentID.String(), value.Owner, fmt.Sprint(value.OwnerGeneration), value.Contract,
		value.ContentSHA256, value.Action, value.CompletedAt.Format(time.RFC3339Nano)}, "\x00")
	digest := sha256.Sum256([]byte(document))
	return hex.EncodeToString(digest[:])
}

func scanOwnerReceipt(source row) (catalogdomain.OwnerReceipt, error) {
	var receiptID, operationID, componentID string
	var value catalogdomain.OwnerReceipt
	if err := source.Scan(&receiptID, &operationID, &componentID, &value.Owner,
		&value.OwnerGeneration, &value.Contract, &value.ContentSHA256,
		&value.Action, &value.Ready, &value.CompletedAt); err != nil {
		return catalogdomain.OwnerReceipt{}, err
	}
	value.CompletedAt = value.CompletedAt.UTC()
	var err error
	value.ReceiptID, err = packagecontract.ParseStableID(receiptID)
	if err != nil {
		return catalogdomain.OwnerReceipt{}, unavailable()
	}
	value.OperationID, err = packagecontract.ParseStableID(operationID)
	if err != nil {
		return catalogdomain.OwnerReceipt{}, unavailable()
	}
	value.ComponentID, err = packagecontract.ParseComponentID(componentID)
	if err != nil || !canonicalTime(value.CompletedAt) || !value.Ready {
		return catalogdomain.OwnerReceipt{}, unavailable()
	}
	return value, nil
}
