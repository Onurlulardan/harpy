package billingv1adoptionapp

type Issue struct {
	ScopeKind string `json:"scope_kind"`
	ScopeID   string `json:"scope_id"`
	Category  string `json:"category"`
	Code      string `json:"code"`
	SourceKey string `json:"source_key"`
}

func issue(scopeKind, scopeID, category, code, sourceKey string) Issue {
	return Issue{
		ScopeKind: scopeKind, ScopeID: scopeID, Category: category,
		Code: code, SourceKey: sourceKey,
	}
}
