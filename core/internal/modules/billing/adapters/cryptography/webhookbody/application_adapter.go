package webhookbody

import (
	"context"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
)

type ApplicationProtector struct {
	sealer *Sealer
}

func NewApplicationProtector(sealer *Sealer) (*ApplicationProtector, error) {
	if sealer == nil {
		return nil, ErrConfiguration
	}
	return &ApplicationProtector{sealer: sealer}, nil
}

func (protector *ApplicationProtector) Protect(
	ctx context.Context,
	associatedData []byte,
	body []byte,
) (providerapp.ProtectedBody, error) {
	if protector == nil || protector.sealer == nil {
		return providerapp.ProtectedBody{}, ErrConfiguration
	}
	envelope, err := protector.sealer.Seal(ctx, associatedData, body)
	if err != nil {
		return providerapp.ProtectedBody{}, err
	}
	return providerapp.NewProtectedBody(
		envelope.Algorithm, envelope.Nonce, envelope.Ciphertext,
		envelope.WrappedDEK.KeyID, envelope.WrappedDEK.Ciphertext,
		envelope.BodySHA256,
	)
}

type ApplicationOpener struct {
	opener *ReconciliationOpener
}

func NewApplicationOpener(opener *ReconciliationOpener) (*ApplicationOpener, error) {
	if opener == nil {
		return nil, ErrConfiguration
	}
	return &ApplicationOpener{opener: opener}, nil
}

func (opener *ApplicationOpener) OpenForReconciliation(
	ctx context.Context,
	associatedData []byte,
	body providerapp.ProtectedBody,
) ([]byte, error) {
	if opener == nil || opener.opener == nil {
		return nil, ErrConfiguration
	}
	return opener.opener.Open(ctx, associatedData, Envelope{
		Algorithm: body.Algorithm(), Nonce: body.Nonce(), Ciphertext: body.Ciphertext(),
		WrappedDEK: WrappedDEK{KeyID: body.KeyID(), Ciphertext: body.WrappedDEK()},
		BodySHA256: body.BodySHA256(),
	})
}
