package notifications

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
)

type deliveryReference struct{ tenantID, deliveryID application.ID }

func parseDeliveryReference(value string) (deliveryReference, error) {
	const prefix = "notification-delivery://tenant/"
	if !strings.HasPrefix(value, prefix) {
		return deliveryReference{}, fmt.Errorf("notification delivery reference is invalid")
	}
	parts := strings.Split(strings.TrimPrefix(value, prefix), "/")
	if len(parts) != 3 || parts[1] != "delivery" {
		return deliveryReference{}, fmt.Errorf("notification delivery reference is invalid")
	}
	tenantID, tenantErr := application.ParseID(parts[0])
	deliveryID, deliveryErr := application.ParseID(parts[2])
	if tenantErr != nil || deliveryErr != nil {
		return deliveryReference{}, fmt.Errorf("notification delivery reference is invalid")
	}
	return deliveryReference{tenantID: tenantID, deliveryID: deliveryID}, nil
}

func deliveryReferenceDigest(reference string) string {
	digest := sha256.Sum256([]byte("notifications.delivery.v1\x00" + reference))
	return hex.EncodeToString(digest[:])
}
