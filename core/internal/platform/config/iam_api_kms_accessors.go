package config

func (config IAMAPI) TOTPKMSKeyID() (IAMKMSKeyID, error) {
	if err := config.Validate(); err != nil {
		return IAMKMSKeyID{}, err
	}
	return config.totpKMS, nil
}

func (config IAMAPI) DeliveryKMSKeyID() (IAMKMSKeyID, error) {
	if err := config.Validate(); err != nil {
		return IAMKMSKeyID{}, err
	}
	return config.deliveryKMS, nil
}
