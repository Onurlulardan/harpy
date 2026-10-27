package authentication

type RotateRequest struct {
	Presented PresentedRefresh
	Client    ClientID
}

func (request RotateRequest) validate() error {
	if nilInterface(request.Presented) ||
		(request.Client != ClientOperatorConsole &&
			request.Client != ClientPortal) {
		return ErrAuthenticationContract
	}
	return nil
}
