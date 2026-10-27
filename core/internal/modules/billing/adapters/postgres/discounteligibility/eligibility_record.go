package discounteligibility

import (
	"time"

	"github.com/google/uuid"
)

type eligibilityRecord struct {
	tenantID                            uuid.UUID
	candidateCount                      int64
	discountID, discountVersionID       uuid.UUID
	subscriptionID, offeringID, priceID uuid.UUID
	revision, version                   int64
	subscriptionCurrency, code, kind    string
	fixedMinorUnits, registryVersion    *int64
	fixedCurrency                       *string
	basisPoints                         *int32
	duration                            string
	repeatingCycles                     *int64
	priority                            int32
	globalLimit, tenantLimit            *int64
	allOfferings                        bool
	expectedOfferingCount               int16
	validStart                          time.Time
	validEnd                            *time.Time
	available                           bool
	globalUsed, tenantUsed              int64
	offeringAllowed, alreadyRedeemed    bool
}

func (record eligibilityRecord) eligible() bool {
	return record.candidateCount == 1 && record.globalUsed >= 0 && record.tenantUsed >= 0 &&
		record.available && record.offeringAllowed && !record.alreadyRedeemed &&
		(record.globalLimit == nil || record.globalUsed < *record.globalLimit) &&
		(record.tenantLimit == nil || record.tenantUsed < *record.tenantLimit) &&
		(record.kind != "fixed" || record.fixedCurrency != nil &&
			*record.fixedCurrency == record.subscriptionCurrency)
}
