package profile

type Classification string

const (
	ClassificationPublic       Classification = "public"
	ClassificationInternal     Classification = "internal"
	ClassificationConfidential Classification = "confidential"
	ClassificationRestricted   Classification = "restricted"
)

func (classification Classification) Valid() bool {
	switch classification {
	case ClassificationPublic, ClassificationInternal,
		ClassificationConfidential, ClassificationRestricted:
		return true
	default:
		return false
	}
}
