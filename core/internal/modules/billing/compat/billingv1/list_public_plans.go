package billingv1

import (
	"net/http"
)

type comparisonFeature struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	LabelEN string `json:"label_en"`
	Type    string `json:"type"`
}

func (handler *Handler) listPublicPlans(w http.ResponseWriter, r *http.Request) {
	offerings, err := handler.publicOfferings(r.Context())
	if err != nil {
		handler.fail(w, "billing.route.public-plans", err)
		return
	}
	plans, err := handler.presentPublicPlans(r.Context(), offerings)
	if err != nil {
		handler.fail(w, "billing.route.public-plans", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"plans": plans, "currency": "TRY", "currency_symbol": "₺",
		"default_billing_interval": "monthly",
		"comparison_features":      legacyComparisonFeatures(),
	})
}

func legacyComparisonFeatures() []comparisonFeature {
	return []comparisonFeature{
		{"max_users", "Kullanıcı Limiti", "User Limit", "number"},
		{"submissions_per_month", "Aylık Kayıt", "Monthly Submissions", "number"},
		{"max_storage_mb", "Depolama", "Storage", "storage"},
		{"max_forms", "Form Sayısı", "Form Count", "number"},
		{"workflows", "İş Akışları", "Workflows", "boolean"},
		{"ai_chat", "AI Sohbet", "AI Chat", "boolean"},
		{"api_access", "API Erişimi", "API Access", "boolean"},
		{"custom_domain", "Özel Domain", "Custom Domain", "boolean"},
	}
}
