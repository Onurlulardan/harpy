package password

const (
	AlgorithmArgon2idV1 = "argon2id_v1"
	MemoryKiBV1         = uint32(65536)
	IterationsV1        = uint32(3)
	ParallelismV1       = uint8(4)
	SaltBytesV1         = 16
	TagBytesV1          = 32
)

type Argon2idCredential struct {
	salt        [SaltBytesV1]byte
	tag         [TagBytesV1]byte
	memoryKiB   uint32
	iterations  uint32
	parallelism uint8
	issued      bool
}

func Argon2idCredentialFromStorage(
	salt []byte,
	tag []byte,
	memoryKiB uint32,
	iterations uint32,
	parallelism uint8,
) (Argon2idCredential, error) {
	if len(salt) != SaltBytesV1 || len(tag) != TagBytesV1 ||
		memoryKiB != MemoryKiBV1 || iterations != IterationsV1 ||
		parallelism != ParallelismV1 {
		return Argon2idCredential{}, ErrCredentialInvalid
	}
	credential := Argon2idCredential{
		memoryKiB: memoryKiB, iterations: iterations,
		parallelism: parallelism, issued: true,
	}
	copy(credential.salt[:], salt)
	copy(credential.tag[:], tag)
	return credential, nil
}

func (credential Argon2idCredential) Algorithm() string       { return AlgorithmArgon2idV1 }
func (credential Argon2idCredential) Salt() [SaltBytesV1]byte { return credential.salt }
func (credential Argon2idCredential) Tag() [TagBytesV1]byte   { return credential.tag }
func (credential Argon2idCredential) MemoryKiB() uint32       { return credential.memoryKiB }
func (credential Argon2idCredential) Iterations() uint32      { return credential.iterations }
func (credential Argon2idCredential) Parallelism() uint8      { return credential.parallelism }

func (credential Argon2idCredential) Valid() bool {
	return credential.issued && credential.memoryKiB == MemoryKiBV1 &&
		credential.iterations == IterationsV1 && credential.parallelism == ParallelismV1
}

func (credential Argon2idCredential) NeedsRehash() bool {
	return !credential.Valid()
}
