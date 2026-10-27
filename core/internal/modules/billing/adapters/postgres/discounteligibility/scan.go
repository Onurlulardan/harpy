package discounteligibility

func scanEligibility(scanner row) (eligibilityRecord, error) {
	var record eligibilityRecord
	err := scanner.Scan(
		&record.tenantID, &record.candidateCount,
		&record.discountID, &record.revision,
		&record.discountVersionID, &record.version,
		&record.subscriptionID, &record.offeringID,
		&record.priceID, &record.subscriptionCurrency,
		&record.code, &record.kind, &record.fixedMinorUnits,
		&record.fixedCurrency, &record.registryVersion,
		&record.basisPoints, &record.duration, &record.repeatingCycles,
		&record.priority, &record.globalLimit, &record.tenantLimit,
		&record.allOfferings, &record.expectedOfferingCount,
		&record.validStart, &record.validEnd, &record.available,
		&record.globalUsed, &record.tenantUsed,
		&record.offeringAllowed, &record.alreadyRedeemed,
	)
	return record, err
}
