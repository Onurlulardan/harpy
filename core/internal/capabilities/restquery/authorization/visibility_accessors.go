package authorization

import "crypto/sha256"

func (visibility Visibility) ResourceRequirementSHA256() ([sha256.Size]byte, error) {
	if err := visibility.Validate(); err != nil {
		return [sha256.Size]byte{}, err
	}
	return visibility.resourceRequirement, nil
}

func (visibility Visibility) Mode() (VisibilityMode, error) {
	if err := visibility.Validate(); err != nil {
		return 0, err
	}
	return visibility.mode, nil
}

func (visibility Visibility) Filter() (MandatoryFilter, bool, error) {
	if err := visibility.Validate(); err != nil {
		return MandatoryFilter{}, false, err
	}
	if visibility.filter == nil {
		return MandatoryFilter{}, false, nil
	}
	return visibility.filter.clone(), true, nil
}

func (visibility Visibility) clone() Visibility {
	result := visibility
	if visibility.filter != nil {
		owned := visibility.filter.clone()
		result.filter = &owned
	}
	return result
}
