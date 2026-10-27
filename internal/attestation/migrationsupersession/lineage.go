package migrationsupersession

func (catalog Catalog) Lineage(terminalPath string) []Entry {
	byReplacement := make(map[string]Entry, len(catalog.entries))
	for _, entry := range catalog.entries {
		byReplacement[entry.replacementPath] = entry
	}
	current, exists := byReplacement[terminalPath]
	if !exists {
		return nil
	}
	reversed := make([]Entry, 0, len(catalog.entries))
	for {
		reversed = append(reversed, current)
		previous, found := byReplacement[current.oldPath]
		if !found {
			break
		}
		current = previous
	}
	lineage := make([]Entry, len(reversed))
	for index := range reversed {
		lineage[len(reversed)-1-index] = reversed[index]
	}
	return lineage
}
