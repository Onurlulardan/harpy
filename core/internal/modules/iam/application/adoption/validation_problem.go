package adoption

type Issue struct {
	Code      string `json:"code"`
	Category  string `json:"category"`
	SourceKey string `json:"source_key"`
}

func issue(code, category, sourceKey string) Issue {
	return Issue{Code: code, Category: category, SourceKey: sourceKey}
}
