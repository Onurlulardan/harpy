package migrationsupersession

import "fmt"

func validateChains(
	contractVersion string,
	chains map[string]map[int]map[string]Entry,
) error {
	maximumLevels := contractDepth(contractVersion)
	highestLevel := 0
	for _, chain := range chains {
		chainDepth := len(chain)
		if chainDepth == 0 || chainDepth > maximumLevels {
			return fmt.Errorf("migration supersession chain depth differs")
		}
		for level := 1; level <= chainDepth; level++ {
			pair, exists := chain[level]
			if !exists {
				return fmt.Errorf("migration supersession chain level is missing")
			}
			if err := validatePair(pair); err != nil {
				return err
			}
		}
		for level := 1; level < chainDepth; level++ {
			for _, origin := range []string{"fresh", "v1_adopted"} {
				first, firstOK := chain[level][origin]
				second, secondOK := chain[level+1][origin]
				if firstOK != secondOK || (firstOK &&
					(first.replacementPath != second.oldPath ||
						first.replacementSHA256 != second.oldSHA256)) {
					return fmt.Errorf("migration supersession chain link differs")
				}
			}
		}
		if chainDepth > highestLevel {
			highestLevel = chainDepth
		}
	}
	if highestLevel != maximumLevels {
		return fmt.Errorf("migration supersession catalog depth differs")
	}
	return nil
}

func validatePair(pair map[string]Entry) error {
	fresh, hasFresh := pair["fresh"]
	adopted, hasAdopted := pair["v1_adopted"]
	if !hasFresh || (hasAdopted && len(pair) != 2) || (!hasAdopted && len(pair) != 1) {
		return fmt.Errorf("migration supersession origin pair differs")
	}
	if hasAdopted && (fresh.oldSHA256 != adopted.oldSHA256 ||
		fresh.replacementSHA256 != adopted.replacementSHA256) {
		return fmt.Errorf("migration supersession origin pair differs")
	}
	return nil
}
