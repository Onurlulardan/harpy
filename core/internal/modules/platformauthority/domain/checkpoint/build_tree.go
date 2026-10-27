package checkpoint

import (
	"fmt"
	"sort"
)

func Build(leaves []ScopeLeaf) (Tree, error) {
	if len(leaves) == 0 || len(leaves) > 1_000_000 {
		return Tree{}, fmt.Errorf("protected authority checkpoint leaf count is invalid")
	}
	ordered := append([]ScopeLeaf(nil), leaves...)
	sort.Slice(ordered, func(left, right int) bool {
		return leafKey(ordered[left]) < leafKey(ordered[right])
	})
	hashes := make([][32]byte, len(ordered))
	for index, leaf := range ordered {
		if index > 0 && leafKey(leaf) == leafKey(ordered[index-1]) {
			return Tree{}, fmt.Errorf("protected authority checkpoint scope is duplicated")
		}
		hash, err := hashLeaf(leaf)
		if err != nil {
			return Tree{}, err
		}
		hashes[index] = hash
	}
	proofs := make([]InclusionProof, len(ordered))
	for index, leaf := range ordered {
		proofs[index] = InclusionProof{Leaf: leaf, LeafIndex: index, LeafCount: len(ordered)}
	}
	root := buildLevels(hashes, proofs)
	return Tree{RootSHA256: digestString(root), Leaves: ordered, proofs: proofs}, nil
}

func buildLevels(level [][32]byte, proofs []InclusionProof) [32]byte {
	positions := make([]int, len(level))
	for index := range positions {
		positions[index] = index
	}
	for len(level) > 1 {
		for leafIndex, position := range positions {
			sibling := position ^ 1
			if sibling >= len(level) {
				sibling = position
			}
			proofs[leafIndex].Siblings = append(
				proofs[leafIndex].Siblings, digestString(level[sibling]),
			)
			positions[leafIndex] /= 2
		}
		next := make([][32]byte, 0, (len(level)+1)/2)
		for index := 0; index < len(level); index += 2 {
			right := index + 1
			if right == len(level) {
				right = index
			}
			next = append(next, hashNode(level[index], level[right]))
		}
		level = next
	}
	return level[0]
}

func (tree Tree) Proof(index int) (InclusionProof, error) {
	if index < 0 || index >= len(tree.proofs) {
		return InclusionProof{}, fmt.Errorf("checkpoint proof index is invalid")
	}
	proof := tree.proofs[index]
	proof.Siblings = append([]string(nil), proof.Siblings...)
	return proof, nil
}
