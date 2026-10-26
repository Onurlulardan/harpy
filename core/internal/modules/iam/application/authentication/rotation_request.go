package authentication

type RotateRequest struct {
	Presented PresentedRefresh
	Client    ClientID
}

func (request RotateRequest) validate() error {
	if nilInterface(request.Presented) ||
		(request.Client != ClientOperatorConsole &&
			request.Client != ClientPortal && request.Client != ClientMCP) {
		return ErrAuthenticationContract
	}
	return nil
}
