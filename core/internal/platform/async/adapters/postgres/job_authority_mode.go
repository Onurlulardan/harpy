package postgres

type jobAuthorityMode string

const (
	jobAuthorityModeV2            jobAuthorityMode = "v2"
	jobAuthorityModeCutoverShadow jobAuthorityMode = "cutover_shadow"
)

func (mode jobAuthorityMode) valid() bool {
	return mode == jobAuthorityModeV2 || mode == jobAuthorityModeCutoverShadow
}
