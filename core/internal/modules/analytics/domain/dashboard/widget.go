package dashboard

import "bytes"

type Widget struct {
	id             ID
	queryVersionID [16]byte
	position       uint16
	parameterHash  [32]byte
	outputHash     [32]byte
}

func NewWidget(
	id ID, queryVersionID [16]byte, position uint16,
	parameterHash, outputHash [32]byte,
) (Widget, error) {
	if !id.Valid() || !validV7(queryVersionID) || position >= 30 ||
		zeroDigest(parameterHash) || zeroDigest(outputHash) {
		return Widget{}, ErrInvalid
	}
	return Widget{id: id, queryVersionID: queryVersionID, position: position,
		parameterHash: parameterHash, outputHash: outputHash}, nil
}

func (value Widget) ID() ID                    { return value.id }
func (value Widget) QueryVersionID() [16]byte  { return value.queryVersionID }
func (value Widget) Position() uint16          { return value.position }
func (value Widget) ParameterSHA256() [32]byte { return value.parameterHash }
func (value Widget) OutputSHA256() [32]byte    { return value.outputHash }

func zeroDigest(value [32]byte) bool {
	return bytes.Equal(value[:], make([]byte, 32))
}
