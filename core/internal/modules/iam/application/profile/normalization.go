package profile

type NormalizationProfile string

const (
	NormalizationNoneV1  NormalizationProfile = "none_v1"
	NormalizationEmailV1 NormalizationProfile = "email_v1"
)

func (profile NormalizationProfile) Valid() bool {
	return profile == NormalizationNoneV1 || profile == NormalizationEmailV1
}
