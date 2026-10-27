package cursor

import "io"

func (codec Codec) nextNonce() ([nonceBytes]byte, error) {
	var nonce [nonceBytes]byte
	if _, err := io.ReadFull(codec.state.nonceSource, nonce[:]); err != nil {
		return [nonceBytes]byte{}, errEncodeFailed
	}
	return nonce, nil
}
