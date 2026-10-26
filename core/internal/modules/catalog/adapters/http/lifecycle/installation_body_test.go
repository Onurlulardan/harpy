package lifecyclehttp

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func installationBody(operation string) (*strings.Reader, error) {
	identifier, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	switch operation {
	case "install":
		return strings.NewReader(fmt.Sprintf(`{"package_id":%q,"version":"1.2.3"}`,
			"e3000000-0000-0000-0000-000000000001")), nil
	case "upgrade":
		return strings.NewReader(`{"version":"1.2.3"}`), nil
	case "purge":
		second, secondErr := uuid.NewV7()
		third, thirdErr := uuid.NewV7()
		if secondErr != nil || thirdErr != nil {
			return nil, fmt.Errorf("create purge identities")
		}
		return strings.NewReader(fmt.Sprintf(
			`{"operation_id":%q,"backup_receipt_id":%q,"approval_receipt_id":%q,`+
				`"reason":"verified","confirmation_slug":"package"}`,
			identifier.String(), second.String(), third.String())), nil
	case "approve_purge":
		second, secondErr := uuid.NewV7()
		third, thirdErr := uuid.NewV7()
		if secondErr != nil || thirdErr != nil {
			return nil, fmt.Errorf("create purge approval identities")
		}
		return strings.NewReader(fmt.Sprintf(
			`{"operation_id":%q,"backup_receipt_id":%q,"requested_by":%q,`+
				`"reason":"verified","confirmation_slug":"package",`+
				`"purge_idempotency_key":"catalog-purge-key-0001"}`,
			identifier.String(), second.String(), third.String())), nil
	default:
		return nil, nil
	}
}

func installationStatus(operation string) int {
	switch operation {
	case "install", "upgrade", "restore", "retire", "purge":
		return http.StatusAccepted
	case "approve_purge":
		return http.StatusCreated
	default:
		return http.StatusOK
	}
}
