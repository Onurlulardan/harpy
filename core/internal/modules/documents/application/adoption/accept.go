package adoption

import "context"

func (application *Application) Accept(ctx context.Context, authority Authority,
	digest [32]byte, finalLSN string,
) (Run, error) {
	if application == nil || ctx == nil || ctx.Err() != nil ||
		!validAuthority(authority) || digest == [32]byte{} || !validLSN(finalLSN) {
		return Run{}, ErrInvalid
	}
	result, err := application.store.Accept(ctx, authority, digest, finalLSN,
		application.clock.Now().UTC())
	if err != nil {
		return Run{}, mapError(err)
	}
	if !validRun(result, authority) || result.State != "accepted" ||
		result.AcceptanceSHA256 != digest || result.FinalLSN != finalLSN ||
		result.AppliedLSN != finalLSN {
		return Run{}, ErrUnavailable
	}
	return result, nil
}
