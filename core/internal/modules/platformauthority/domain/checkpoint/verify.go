package checkpoint

import "fmt"

func VerifyInclusion(proof InclusionProof, expectedRoot string) error {
	if proof.LeafCount < 1 || proof.LeafIndex < 0 || proof.LeafIndex >= proof.LeafCount ||
		!checkpointDigestPattern.MatchString(expectedRoot) {
		return fmt.Errorf("protected authority checkpoint proof is invalid")
	}
	current, err := hashLeaf(proof.Leaf)
	if err != nil {
		return err
	}
	position := proof.LeafIndex
	width := proof.LeafCount
	for _, encodedSibling := range proof.Siblings {
		if width <= 1 {
			return fmt.Errorf("protected authority checkpoint proof is oversized")
		}
		sibling, err := parseDigest(encodedSibling)
		if err != nil {
			return err
		}
		if position%2 == 0 {
			current = hashNode(current, sibling)
		} else {
			current = hashNode(sibling, current)
		}
		position /= 2
		width = (width + 1) / 2
	}
	if width != 1 || digestString(current) != expectedRoot {
		return fmt.Errorf("protected authority checkpoint proof mismatches")
	}
	return nil
}

func VerifyExactSnapshot(leaves []ScopeLeaf, expectedRoot string, expectedCount int64) error {
	if int64(len(leaves)) != expectedCount {
		return fmt.Errorf("protected authority checkpoint scope count mismatches")
	}
	tree, err := Build(leaves)
	if err != nil {
		return err
	}
	if tree.RootSHA256 != expectedRoot {
		return fmt.Errorf("protected authority checkpoint snapshot mismatches")
	}
	return nil
}
