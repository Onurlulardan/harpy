package licensev2

type DerivedOffering struct {
	Schema                 string         `json:"schema"`
	OfferingVersionID      string         `json:"offering_version_id"`
	Revision               int64          `json:"revision"`
	EffectiveAt            string         `json:"effective_at"`
	LicenseID              string         `json:"license_id"`
	LicenseeOrganizationID string         `json:"licensee_organization_id"`
	EnvironmentID          string         `json:"environment_id"`
	Issuer                 string         `json:"issuer"`
	SnapshotSequence       int64          `json:"snapshot_sequence"`
	SnapshotSHA256         string         `json:"snapshot_sha256"`
	ReleaseChannel         string         `json:"release_channel"`
	Booleans               []BooleanGrant `json:"booleans"`
	Limits                 []LimitGrant   `json:"limits"`
}

type BooleanGrant struct {
	Key   string `json:"key"`
	Value bool   `json:"value"`
}

type LimitGrant struct {
	Key       string `json:"key"`
	Value     int64  `json:"value"`
	Unlimited bool   `json:"unlimited"`
}
