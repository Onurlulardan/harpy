package actorassertionv1

type Domain string

const (
	DomainAnalytics     Domain = "analytics"
	DomainBilling       Domain = "billing"
	DomainCatalog       Domain = "catalog"
	DomainDocuments     Domain = "documents"
	DomainForms         Domain = "forms"
	DomainNotifications Domain = "notifications"
	DomainWork          Domain = "work"
	DomainWorkflow      Domain = "workflow"
)

func (domain Domain) Validate() error {
	switch domain {
	case DomainAnalytics, DomainBilling, DomainCatalog, DomainDocuments,
		DomainForms, DomainNotifications, DomainWork, DomainWorkflow:
		return nil
	default:
		return ErrContract
	}
}
