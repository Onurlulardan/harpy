package systemresources

import publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"

type MutationResult struct {
	mutation     publicresource.MutationResult
	presentation Presentation
	iamOnly      bool
	validated    bool
}

func NewMutationResult(
	mutation publicresource.MutationResult,
	presentation Presentation,
) (MutationResult, error) {
	result := MutationResult{
		mutation: mutation, presentation: presentation, validated: true,
	}
	if result.Validate() != nil {
		return MutationResult{}, ErrContract
	}
	return result, nil
}

func newIAMOnlyMutationResult(
	mutation publicresource.MutationResult,
) (MutationResult, error) {
	result := MutationResult{mutation: mutation, iamOnly: true, validated: true}
	if result.Validate() != nil {
		return MutationResult{}, ErrContract
	}
	return result, nil
}

func (result MutationResult) Validate() error {
	if !result.validated || result.mutation.Validate() != nil {
		return ErrContract
	}
	if result.iamOnly {
		return nil
	}
	read, err := result.mutation.Read()
	if err != nil {
		return ErrContract
	}
	_, err = NewReadResult(read, result.presentation)
	return err
}

func (result MutationResult) IAM() (publicresource.MutationResult, error) {
	if result.Validate() != nil {
		return publicresource.MutationResult{}, ErrContract
	}
	return result.mutation, nil
}

func (result MutationResult) Presentation() (Presentation, error) {
	if result.iamOnly {
		return Presentation{}, ErrUnavailable
	}
	if result.Validate() != nil {
		return Presentation{}, ErrContract
	}
	return result.presentation, nil
}
