package documents

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

type platformActorBilling struct{ service ConfiguredDocumentsBilling }
type platformArchiveBilling struct{ service ConfiguredDocumentsBilling }
type platformDeletionBilling struct{ service ConfiguredDocumentsBilling }
type platformReconcileBilling struct{ service ConfiguredDocumentsBilling }

type platformBilling struct {
	actor     platformActorBilling
	archive   platformArchiveBilling
	deletion  platformDeletionBilling
	reconcile platformReconcileBilling
}

func newPlatformBilling(service ConfiguredDocumentsBilling) (platformBilling, error) {
	if foundation.NilPort(service) {
		return platformBilling{}, fmt.Errorf("Documents Billing usage is required")
	}
	return platformBilling{actor: platformActorBilling{service: service},
		archive:   platformArchiveBilling{service: service},
		deletion:  platformDeletionBilling{service: service},
		reconcile: platformReconcileBilling{service: service}}, nil
}
