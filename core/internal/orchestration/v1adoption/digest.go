package v1adoption

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

func digest(value any) string {
	document, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	result := sha256.Sum256(document)
	return hex.EncodeToString(result[:])
}

func validSHA(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size
}

func sortIssues(issues []Issue) {
	sort.Slice(issues, func(i, j int) bool {
		left := issues[i].Code + issues[i].Tenant + issues[i].Evidence
		right := issues[j].Code + issues[j].Tenant + issues[j].Evidence
		return left < right
	})
}
