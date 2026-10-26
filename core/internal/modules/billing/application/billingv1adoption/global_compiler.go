package billingv1adoptionapp

import (
	"sort"

	"github.com/google/uuid"
)

func compileGlobal(
	snapshot GlobalSnapshot,
	mapping AdoptionMapping,
) (ScopePlan, GlobalIndex, error) {
	plan := emptyScopePlan("global", snapshot.EnvironmentID.String(),
		mapping.SourceTimezone, snapshot.SnapshotAt)
	index := newGlobalIndex()
	plans := append([]SourcePlan(nil), snapshot.Plans...)
	campaigns := append([]SourceCampaign(nil), snapshot.Campaigns...)
	coupons := append([]SourceCoupon(nil), snapshot.Coupons...)
	sort.Slice(plans, func(i, j int) bool { return plans[i].ID.String() < plans[j].ID.String() })
	sort.Slice(campaigns, func(i, j int) bool { return campaigns[i].ID.String() < campaigns[j].ID.String() })
	sort.Slice(coupons, func(i, j int) bool { return coupons[i].ID.String() < coupons[j].ID.String() })
	for _, category := range []struct {
		name  string
		rows  any
		count int
	}{
		{"plans", plans, len(plans)}, {"campaigns", campaigns, len(campaigns)},
		{"coupons", coupons, len(coupons)},
	} {
		evidence, err := categoryEvidence(category.name, category.rows, category.count)
		if err != nil {
			return ScopePlan{}, GlobalIndex{}, err
		}
		plan.Categories = append(plan.Categories, evidence)
	}
	var err error
	plan.Offerings, index, plan.Issues = compileOfferings(
		plans, mapping.Global, snapshot.EnvironmentID, index, plan.Issues,
	)
	plan.Discounts, index, plan.Issues = compileDiscounts(
		coupons, mapping.Global, index, plan.Issues, snapshot.EnvironmentID,
	)
	plan.Campaigns, plan.Issues = compileCampaigns(
		campaigns, coupons, index, plan.Issues, snapshot.EnvironmentID,
	)
	plan.ManualProviderAccountID = index.ManualProviderAccountID
	providerEvidence, evidenceErr := categoryEvidence(
		"provider_accounts", []string{"manual/default"}, 1,
	)
	if evidenceErr != nil {
		return ScopePlan{}, GlobalIndex{}, evidenceErr
	}
	plan.Categories = append(plan.Categories, providerEvidence)
	plan.SourceSHA256, err = sourceRoot(plan.Categories)
	if err != nil {
		return ScopePlan{}, GlobalIndex{}, err
	}
	if _, exists := index.Offerings[mapping.Global.BasePlanID]; !exists ||
		mapping.Global.BasePlanID == uuid.Nil {
		plan.Issues = append(plan.Issues, issue("global", snapshot.EnvironmentID.String(),
			"plans", "base_plan_missing", mapping.Global.BasePlanID.String()))
	} else {
		index.Base = index.Offerings[mapping.Global.BasePlanID]
	}
	finalizePlan(&plan)
	return plan, index, nil
}
