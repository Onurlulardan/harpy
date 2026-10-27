package config

func resolveBucketOwner(
	values environmentValues,
	key string,
	environment Environment,
	pathStyle bool,
) (string, bool, error) {
	owner, found := values.lookup(key)
	if !found || owner == "" {
		if environment.productionLike() || !pathStyle {
			return "", false, configError(key, "must be a 12-digit owner ID")
		}
		return "", true, nil
	}
	if len(owner) != 12 || !decimalDigits(owner) {
		return "", false, configError(key, "must be a 12-digit owner ID")
	}
	return owner, false, nil
}

func bucketOwnerAccepted(owner string, optional bool) bool {
	if optional {
		return owner == ""
	}
	return len(owner) == 12 && decimalDigits(owner)
}
