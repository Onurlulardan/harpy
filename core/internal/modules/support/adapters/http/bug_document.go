package supporthttp

import (
	"encoding/json"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

type bugCreateDocument struct {
	Title            string             `json:"title"`
	Description      string             `json:"description"`
	Category         domain.BugCategory `json:"category"`
	Severity         domain.BugSeverity `json:"severity"`
	StepsToReproduce string             `json:"steps_to_reproduce"`
	PageURL          string             `json:"page_url"`
	UserAgent        string             `json:"user_agent"`
	BrowserInfo      json.RawMessage    `json:"browser_info"`
	ConsoleErrors    json.RawMessage    `json:"console_errors"`
	NetworkInfo      json.RawMessage    `json:"network_info"`
}

type bugUpdateDocument struct {
	Status           *domain.BugStatus   `json:"status"`
	Priority         *domain.BugPriority `json:"priority"`
	Resolution       *string             `json:"resolution"`
	AdminNotes       *string             `json:"admin_notes"`
	ExpectedRevision int64               `json:"expected_revision"`
}

func bugInput(document bugCreateDocument) domain.BugInput {
	return domain.BugInput{Title: document.Title, Description: document.Description,
		Category: document.Category, Severity: document.Severity,
		StepsToReproduce: document.StepsToReproduce, PageURL: document.PageURL,
		UserAgent: document.UserAgent, BrowserInfo: document.BrowserInfo,
		ConsoleErrors: document.ConsoleErrors, NetworkInfo: document.NetworkInfo}
}

func decodeBugUpdate(response http.ResponseWriter, request *http.Request,
	id, tenantID domain.ID,
) (domain.BugUpdate, error) {
	var document bugUpdateDocument
	canonical, err := decodeObject(response, request, &document)
	if err != nil {
		return domain.BugUpdate{}, err
	}
	key, digest, err := commandEvidence(request, canonical)
	if err != nil {
		return domain.BugUpdate{}, err
	}
	return domain.BugUpdate{ID: id, TenantID: tenantID, Status: document.Status,
		Priority: document.Priority, Resolution: document.Resolution,
		AdminNotes: document.AdminNotes, ExpectedRevision: document.ExpectedRevision,
		IdempotencyKey: key, RequestSHA256: digest}, nil
}
