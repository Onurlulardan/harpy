//go:build integration

package postgres

import "testing"

func TestIAMV1AdoptionFixtureCoversEveryInventoryCategory(t *testing.T) {
	for _, tenant := range iamFoundationTenants {
		for _, category := range iamV1AdoptionCategories[:14] {
			if len(adoptedTargets(tenant, category)) == 0 {
				t.Fatalf("%s fixture has no %s target", tenant.schema, category)
			}
		}
		if count := adoptedTargetCount(tenant); count != 21 {
			t.Fatalf("%s fixture target count = %d, want 21", tenant.schema, count)
		}
	}
}
