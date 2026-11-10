package authentication

type IssueRequest struct {
	Secret PasswordSecret
	Client ClientID
	Device DeviceEvidence
}

func (request IssueRequest) validate() error {
	if nilInterface(request.Secret) || request.Device.validate() != nil ||
		request.Client != ClientOperatorConsole {
		return ErrAuthenticationContract
	}
	return nil
}
