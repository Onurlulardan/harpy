package reports

type resourceSpec struct {
	Name       string      `json:"name"`
	ResourceID string      `json:"resource_id"`
	VersionID  string      `json:"version_id"`
	View       string      `json:"view"`
	StableID   string      `json:"stable_id"`
	Operation  string      `json:"operation"`
	Aggregate  bool        `json:"aggregate"`
	Fields     []fieldSpec `json:"fields"`
}

func resources() []resourceSpec {
	return []resourceSpec{
		subscriptionSpec(), auditSpec(), legacyEventSpec(), revenueSpec(),
	}
}
