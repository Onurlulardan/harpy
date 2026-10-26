package menu

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"

	menudomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/menu"
)

const emptyMenuDigest = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

func replaceDigest(items []menudomain.Item) string {
	if len(items) == 0 {
		return emptyMenuDigest
	}
	lines := make([]string, 0, len(items))
	for _, item := range items {
		parent, target := "", ""
		if item.ParentID != nil {
			parent = item.ParentID.String()
		}
		if item.TargetID != nil {
			target = item.TargetID.String()
		}
		lines = append(lines, strings.Join([]string{
			item.PackageID.String(), item.ID.String(), parent, string(item.Kind),
			target, item.Label, strconv.Itoa(item.Order),
		}, "\n"))
	}
	sort.Strings(lines)
	digest := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(digest[:])
}
