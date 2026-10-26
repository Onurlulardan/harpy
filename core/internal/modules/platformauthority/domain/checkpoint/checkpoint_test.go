package checkpoint

import (
	"strings"
	"testing"
)

func TestSortedMerkleProofAndExactSnapshot(t *testing.T) {
	leaves := checkpointLeaves()
	tree, err := Build([]ScopeLeaf{leaves[2], leaves[0], leaves[1]})
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.Leaves) != 3 || tree.Leaves[0].ScopeKind != "environment" {
		t.Fatal("checkpoint leaves were not canonical sorted")
	}
	for index := range tree.Leaves {
		proof, err := tree.Proof(index)
		if err != nil {
			t.Fatal(err)
		}
		if err := VerifyInclusion(proof, tree.RootSHA256); err != nil {
			t.Fatalf("proof %d failed: %v", index, err)
		}
	}
	if err := VerifyExactSnapshot(leaves, tree.RootSHA256, 3); err != nil {
		t.Fatal(err)
	}
}

func TestCheckpointRejectsMissingExtraStaleAndTamperedProof(t *testing.T) {
	leaves := checkpointLeaves()
	tree, err := Build(leaves)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyExactSnapshot(leaves[:2], tree.RootSHA256, 3); err == nil {
		t.Fatal("missing active scope was accepted")
	}
	extra := append(append([]ScopeLeaf(nil), leaves...), ScopeLeaf{
		RegistryBindingID: "11111111-1111-4111-8111-111111111111",
		ScopeKind:         "tenant", ScopeID: "55555555-5555-4555-8555-555555555555",
		Generation: 1, StateSHA256: strings.Repeat("d", 64),
		AckSHA256: strings.Repeat("e", 64),
	})
	if err := VerifyExactSnapshot(extra, tree.RootSHA256, 3); err == nil {
		t.Fatal("extra scope was accepted")
	}
	stale := append([]ScopeLeaf(nil), leaves...)
	stale[1].Generation--
	if err := VerifyExactSnapshot(stale, tree.RootSHA256, 3); err == nil {
		t.Fatal("stale scope was accepted")
	}
	proof, err := tree.Proof(1)
	if err != nil {
		t.Fatal(err)
	}
	proof.Siblings[0] = strings.Repeat("f", 64)
	if err := VerifyInclusion(proof, tree.RootSHA256); err == nil {
		t.Fatal("tampered inclusion proof was accepted")
	}
}

func checkpointLeaves() []ScopeLeaf {
	binding := "11111111-1111-4111-8111-111111111111"
	return []ScopeLeaf{
		{RegistryBindingID: binding, ScopeKind: "environment", ScopeID: binding,
			Generation: 3, StateSHA256: strings.Repeat("a", 64), AckSHA256: strings.Repeat("b", 64)},
		{RegistryBindingID: binding, ScopeKind: "tenant",
			ScopeID:    "22222222-2222-4222-8222-222222222222",
			Generation: 2, StateSHA256: strings.Repeat("b", 64), AckSHA256: strings.Repeat("c", 64)},
		{RegistryBindingID: binding, ScopeKind: "tenant",
			ScopeID:    "33333333-3333-4333-8333-333333333333",
			Generation: 1, StateSHA256: strings.Repeat("c", 64), AckSHA256: strings.Repeat("d", 64)},
	}
}
