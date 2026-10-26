package catalogquery

import "testing"

func TestOfferingChildrenDecodeTypedDocuments(t *testing.T) {
	prices, grants, err := decodeOfferingChildren(
		[]byte(`[{"id":"00000000-0000-0000-0000-000000000001","cadence":"monthly","minor_units":9900,"currency":"USD","currency_registry_version":1}]`),
		[]byte(`[{"key":"feature.enabled","kind":"boolean","boolean":true,"integer":null,"unlimited":false}]`),
	)
	if err != nil || len(prices) != 1 || prices[0].Currency != "USD" ||
		len(grants) != 1 || grants[0].Boolean == nil || !*grants[0].Boolean {
		t.Fatalf("decoded offering children = %#v %#v, %v", prices, grants, err)
	}
}

func TestActiveDiscountDecodePreservesTypedPolicy(t *testing.T) {
	discount, err := decodeDiscount([]byte(`{
		"id":"00000000-0000-0000-0000-000000000010",
		"code":"SAVE15","revision":2,
		"version_id":"00000000-0000-0000-0000-000000000011","version":2,
		"kind":"basis_point","basis_points":1500,"fixed_minor_units":null,
		"currency":null,"currency_registry_version":null,"duration":"once",
		"duration_cycles":null,"priority":10,"all_offerings":true,
		"offering_version_ids":[],"valid_start":"2026-07-17T00:00:00Z",
		"valid_end":null,"global_use_limit":100,"per_tenant_use_limit":1
	}`))
	if err != nil || discount.BasisPoints == nil || *discount.BasisPoints != 1500 ||
		discount.GlobalUseLimit == nil || *discount.GlobalUseLimit != 100 ||
		!discount.AllOfferings {
		t.Fatalf("decoded discount = %#v, %v", discount, err)
	}
}
