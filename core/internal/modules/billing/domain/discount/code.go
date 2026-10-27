package discount

const maximumCodeLength = 64

type Code string

func NormalizeCode(value string) (Code, error) {
	if len(value) == 0 || len(value) > maximumCodeLength {
		return "", ErrCode
	}
	normalized := make([]byte, len(value))
	for index := range value {
		character := value[index]
		switch {
		case character >= 'a' && character <= 'z':
			normalized[index] = character - ('a' - 'A')
		case character >= 'A' && character <= 'Z',
			character >= '0' && character <= '9', character == '-', character == '_':
			normalized[index] = character
		default:
			return "", ErrCode
		}
	}
	return Code(normalized), nil
}

func (code Code) Valid() bool {
	normalized, err := NormalizeCode(string(code))
	return err == nil && normalized == code
}
