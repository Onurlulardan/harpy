package lifecycle

import (
	"testing"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
)

func TestOwnerReceiptRootIsOrderIndependent(t *testing.T) {
	first := ownerReceiptRoot([]catalogdomain.OwnerReceipt{{Owner: "iam"}, {Owner: "forms"}})
	second := ownerReceiptRoot([]catalogdomain.OwnerReceipt{{Owner: "forms"}, {Owner: "iam"}})
	if first != second || !digestPattern.MatchString(first) {
		t.Fatalf("receipt roots differ: %s %s", first, second)
	}
}
