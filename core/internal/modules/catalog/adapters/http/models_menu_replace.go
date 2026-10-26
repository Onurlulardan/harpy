package cataloghttp

type MenuReplaceItem struct {
	ItemID   string  `json:"item_id"`
	ParentID *string `json:"parent_id"`
	Kind     string  `json:"kind"`
	Label    string  `json:"label"`
	Order    int     `json:"order"`
	TargetID *string `json:"target_id"`
	Visible  bool    `json:"visible"`
}

type MenuReplaceInput struct {
	ExpectedRevision uint64            `json:"expected_revision"`
	Items            []MenuReplaceItem `json:"items"`
	IdempotencyKey   string
}
