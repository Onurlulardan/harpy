package adoption

import "fmt"

type NormalizeIdentifier func(string) (string, error)

type normalizers struct {
	email  NormalizeIdentifier
	portal NormalizeIdentifier
}

func newNormalizers(email, portal NormalizeIdentifier) (normalizers, error) {
	if email == nil || portal == nil {
		return normalizers{}, fmt.Errorf("IAM V1 adoption normalizers are required")
	}
	return normalizers{email: email, portal: portal}, nil
}
