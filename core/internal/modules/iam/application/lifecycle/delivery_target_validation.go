package lifecycle

func (service *Service) validDeliveryTarget(
	target DeliveryTarget,
	err error,
) (DeliveryTarget, error) {
	if err != nil {
		return DeliveryTarget{}, err
	}
	if target.Validate() != nil {
		return DeliveryTarget{}, ErrUnavailable
	}
	return target, nil
}
