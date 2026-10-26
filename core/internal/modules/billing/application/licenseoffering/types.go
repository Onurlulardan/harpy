package licenseoffering

import (
	"context"
	"time"

	licensebootstrap "github.com/Onurlulardan/harpy/core/internal/capabilities/licensebootstrap/api"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	"github.com/google/uuid"
)

type Plan struct {
	OfferingVersionID uuid.UUID
	Revision          int64
	EffectiveAt       time.Time
}

type Document struct {
	Bytes      []byte
	SHA256     string
	Projection billingapi.LicenseEntitlementProjection
}

type PublishRequest struct {
	Binding  licensebootstrap.Binding
	Offering Document
}

type Receipt struct {
	OperationID       uuid.UUID
	OfferingVersionID uuid.UUID
	Revision          int64
	OfferingSHA256    string
	Replay            bool
}

type Store interface {
	PublishLicenseOffering(context.Context, PublishRequest) (Receipt, error)
}
