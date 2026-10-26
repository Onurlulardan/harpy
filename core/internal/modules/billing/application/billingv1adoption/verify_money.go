package billingv1adoptionapp

func validateTenantMoney(plan *ScopePlan, global GlobalIndex) {
	redemptions := map[string][]RedemptionRecord{}
	for _, redemption := range plan.Redemptions {
		redemptions[redemption.SubscriptionID] = append(
			redemptions[redemption.SubscriptionID], redemption,
		)
	}
	prices := map[string]int64{}
	for _, offering := range global.Offerings {
		prices[offering.MonthlyPriceID] = offering.MonthlyMinorUnits
	}
	for _, subscription := range plan.Subscriptions {
		price, exists := prices[subscription.PriceVersionID]
		if !exists {
			plan.Issues = append(plan.Issues, issue("tenant", plan.ScopeID,
				"money", "price_mapping_missing", subscription.SourceID))
			continue
		}
		applied := redemptions[subscription.ID]
		if len(applied) > 1 {
			plan.Issues = append(plan.Issues, issue("tenant", plan.ScopeID,
				"money", "legacy_last_coupon_wins_ambiguous", subscription.SourceID))
			continue
		}
		expected := price
		if subscription.State == "trialing" && len(applied) == 0 {
			expected = 0
		}
		if len(applied) == 1 {
			expected -= applied[0].SourceDiscountMinorUnits
		}
		if expected < 0 {
			expected = 0
		}
		if expected != subscription.EffectiveMinorUnits {
			plan.Issues = append(plan.Issues, issue("tenant", plan.ScopeID,
				"money", "effective_price_mismatch", subscription.SourceID))
		}
	}
}

func validateCouponCounts(global *ScopePlan, tenants []ScopePlan) {
	actual := map[string]int64{}
	perTenant := map[string]map[string]int64{}
	for _, tenant := range tenants {
		perTenant[tenant.ScopeID] = map[string]int64{}
		for _, redemption := range tenant.Redemptions {
			actual[redemption.DiscountVersionID]++
			perTenant[tenant.ScopeID][redemption.DiscountVersionID]++
		}
	}
	for _, discount := range global.Discounts {
		if actual[discount.VersionID] != discount.SourceUsedCount {
			global.Issues = append(global.Issues, issue("global", global.ScopeID,
				"coupons", "global_use_count_mismatch", discount.SourceID))
		}
		if discount.PerTenantUseLimit == nil {
			continue
		}
		for tenantID, counts := range perTenant {
			if counts[discount.VersionID] > *discount.PerTenantUseLimit {
				global.Issues = append(global.Issues, issue("global", global.ScopeID,
					"coupons", "tenant_use_count_exceeded", tenantID+":"+discount.SourceID))
			}
		}
	}
}
