package invitationdelivery

import deliveryartifactv1 "github.com/Onurlulardan/harpy/contracts/deliveryartifact/v1"

func deliveryReference(value string) (deliveryartifactv1.Reference, error) {
	return deliveryartifactv1.ParseReference(value)
}
