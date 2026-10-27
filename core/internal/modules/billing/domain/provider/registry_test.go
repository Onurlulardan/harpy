package provider

import "testing"

func TestRegistryRequiresManualAndResolvesExactKey(t *testing.T) {
	externalKey, err := NewKey("stripe_connect")
	if err != nil {
		t.Fatal(err)
	}
	external, err := NewDescriptor(externalKey, Capabilities{
		Commands: true, Webhooks: true, Reconciliation: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	registry, err := NewRegistry(ManualDescriptor(), external)
	if err != nil || !registry.Contains(Manual) || !registry.Contains(externalKey) {
		t.Fatalf("registry=%+v, err=%v", registry, err)
	}
	if _, err := registry.Resolve(Key("STRIPE_CONNECT")); err != ErrUnknown {
		t.Fatalf("expected ErrUnknown, got %v", err)
	}
}

func TestRegistryRejectsMissingOrDuplicateManual(t *testing.T) {
	if _, err := NewRegistry(); err != ErrRegistry {
		t.Fatalf("missing: expected ErrRegistry, got %v", err)
	}
	if _, err := NewRegistry(ManualDescriptor(), ManualDescriptor()); err != ErrRegistry {
		t.Fatalf("duplicate: expected ErrRegistry, got %v", err)
	}
}

func TestManualCapabilitiesAreFinal(t *testing.T) {
	if _, err := NewDescriptor(Manual, Capabilities{Commands: true, Webhooks: true,
		Reconciliation: true}); err != ErrDescriptor {
		t.Fatalf("manual webhook: expected ErrDescriptor, got %v", err)
	}
	manual := ManualDescriptor().Capabilities()
	if !manual.Commands || !manual.Reconciliation || manual.Webhooks {
		t.Fatalf("unexpected manual capabilities: %+v", manual)
	}
}

func TestProviderKeyIsStrictLowerASCII(t *testing.T) {
	for _, value := range []string{"stripe", "stripe_connect", "p2"} {
		if _, err := NewKey(value); err != nil {
			t.Fatalf("%q rejected: %v", value, err)
		}
	}
	for _, value := range []string{"", "Stripe", "stripe-connect", "ıyzico", "_stripe"} {
		if _, err := NewKey(value); err != ErrKey {
			t.Fatalf("%q: expected ErrKey, got %v", value, err)
		}
	}
}
