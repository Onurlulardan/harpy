package transaction

import "testing"

func TestProviderLookupCarriesVerifiedObjectWithoutTenantInput(t *testing.T) {
	authentication := [32]byte{1}
	lookup, err := NewProviderLookup(
		"testpay", "acct_10", "subscription", "sub_20", "evt_30", authentication,
	)
	if err != nil {
		t.Fatal(err)
	}
	if lookup.OperationID() != providerWebhookReceiveOperation ||
		lookup.Provider() != "testpay" || lookup.AccountID() != "acct_10" ||
		lookup.ObjectKind() != "subscription" || lookup.ObjectID() != "sub_20" ||
		lookup.EventID() != "evt_30" || ProviderLookupBinding(lookup) == [32]byte{} {
		t.Fatalf("invalid provider lookup: %+v", lookup)
	}
}

func TestProviderLookupRejectsMalformedVerifiedIdentity(t *testing.T) {
	tests := []struct {
		provider, account, kind, object, event string
		digest                                 [32]byte
	}{
		{provider: "TestPay", account: "acct", kind: "subscription", object: "sub", event: "evt", digest: [32]byte{1}},
		{provider: "testpay", account: "bad/account", kind: "subscription", object: "sub", event: "evt", digest: [32]byte{1}},
		{provider: "testpay", account: "acct", kind: "metadata", object: "sub", event: "evt", digest: [32]byte{1}},
		{provider: "testpay", account: "acct", kind: "subscription", object: "sub\n2", event: "evt", digest: [32]byte{1}},
		{provider: "testpay", account: "acct", kind: "subscription", object: "sub", event: "evt"},
	}
	for _, test := range tests {
		if _, err := NewProviderLookup(
			test.provider, test.account, test.kind, test.object, test.event, test.digest,
		); err == nil {
			t.Fatalf("accepted invalid provider lookup: %+v", test)
		}
	}
}
