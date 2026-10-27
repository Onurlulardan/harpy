package installationcredential

type Credential struct {
	salt [16]byte
	tag  [32]byte
}

func NewCredential(salt [16]byte, tag [32]byte) (Credential, error) {
	if salt == [16]byte{} || tag == [32]byte{} {
		return Credential{}, ErrInvalid
	}
	return Credential{salt: salt, tag: tag}, nil
}

func (credential Credential) Valid() bool {
	return credential.salt != [16]byte{} && credential.tag != [32]byte{}
}

func (credential Credential) Salt() [16]byte { return credential.salt }
func (credential Credential) Tag() [32]byte  { return credential.tag }
