package actorassertionv1

type Domain string

const (
	DomainAnalytics     Domain = "analytics"
	DomainBilling       Domain = "billing"
	DomainCatalog       Domain = "catalog"
	DomainDocuments     Domain = "documents"
	DomainNotifications Domain = "notifications"
	DomainWork          Domain = "work"
)

func (domain Domain) Validate() error {
	switch domain {
	case DomainAnalytics, DomainBilling, DomainCatalog, DomainDocuments,
		DomainNotifications, DomainWork:
		return nil
	default:
		return ErrContract
	}
}
