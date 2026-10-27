package publicationhttp

type createDocument struct {
	ResourceKind      string `json:"resource_kind"`
	ResourceVersionID string `json:"resource_version_id"`
	Audience          string `json:"audience"`
	CachePolicy       string `json:"cache_policy"`
}

type transitionDocument struct {
	ExpectedRevision uint64 `json:"expected_revision"`
}

type activateDocument struct {
	ExpectedRevision     uint64 `json:"expected_revision"`
	CapabilityTTLSeconds uint64 `json:"capability_ttl_seconds"`
}
