package eventsv1

import "testing"

func TestConsumerCatalogResolvesOneQueueOwnerAcrossSubscriptions(t *testing.T) {
	first, _ := NewConsumerPolicy(
		"workflow.projector", 1, "workflow", "forms.entry.created", 1,
		"workflow.projector.v1", EffectDatabaseTransactional, 8,
	)
	second, _ := NewConsumerPolicy(
		"workflow.projector", 1, "workflow", "forms.entry.deleted", 1,
		"workflow.projector.v1", EffectDatabaseTransactional, 8,
	)
	catalog, err := NewConsumerCatalog([]ConsumerPolicy{second, first})
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := catalog.ResolveQueue("workflow.projector.v1")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Consumer != "workflow.projector" || resolved.ConsumerVersion != 1 {
		t.Fatalf("queue owner = %#v", resolved)
	}
	if _, err := catalog.ResolveQueue("missing.queue.v1"); err == nil {
		t.Fatal("missing queue resolved")
	}
}
