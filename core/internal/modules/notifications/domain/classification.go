package domain

type Category string

const (
	CategorySecurity       Category = "security"
	CategoryWorkflowAction Category = "workflow_action"
	CategoryBilling        Category = "billing"
	CategoryDocument       Category = "document"
	CategorySystem         Category = "system"
	CategoryProduct        Category = "product"
)

func (value Category) Valid() bool {
	switch value {
	case CategorySecurity, CategoryWorkflowAction, CategoryBilling,
		CategoryDocument, CategorySystem, CategoryProduct:
		return true
	default:
		return false
	}
}

type Channel string

const (
	ChannelInApp    Channel = "in_app"
	ChannelEmail    Channel = "email"
	ChannelTelegram Channel = "telegram"
)

func (value Channel) Valid() bool {
	return value == ChannelInApp || value == ChannelEmail || value == ChannelTelegram
}

func ResolveLocale(requested string, supported []string) (string, bool) {
	for _, candidate := range []string{requested, "tr", "en"} {
		for _, available := range supported {
			if candidate != "" && candidate == available {
				return candidate, true
			}
		}
	}
	return "", false
}
