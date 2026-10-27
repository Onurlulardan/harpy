package systemresources

import publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"

type ReadResult struct {
	read         publicresource.ReadResult
	presentation Presentation
	iamOnly      bool
	validated    bool
}

func NewReadResult(
	read publicresource.ReadResult,
	presentation Presentation,
) (ReadResult, error) {
	result := ReadResult{read: read, presentation: presentation, validated: true}
	if result.Validate() != nil {
		return ReadResult{}, ErrContract
	}
	return result, nil
}

func newIAMOnlyReadResult(read publicresource.ReadResult) (ReadResult, error) {
	result := ReadResult{read: read, iamOnly: true, validated: true}
	if result.Validate() != nil {
		return ReadResult{}, ErrContract
	}
	return result, nil
}

func (result ReadResult) Validate() error {
	if !result.validated || result.read.Validate() != nil {
		return ErrContract
	}
	if result.iamOnly {
		return nil
	}
	if result.presentation.Validate() != nil {
		return ErrContract
	}
	readCatalog, err := result.read.Catalog()
	presentationCatalog, presentationErr := result.presentation.Catalog()
	if err != nil || presentationErr != nil {
		return ErrContract
	}
	equal, err := readCatalog.ExactEqual(presentationCatalog)
	if err != nil || !equal {
		return ErrContract
	}
	return nil
}

func (result ReadResult) IAM() (publicresource.ReadResult, error) {
	if result.Validate() != nil {
		return publicresource.ReadResult{}, ErrContract
	}
	return result.read, nil
}

func (result ReadResult) Presentation() (Presentation, error) {
	if result.iamOnly {
		return Presentation{}, ErrUnavailable
	}
	if result.Validate() != nil {
		return Presentation{}, ErrContract
	}
	return result.presentation, nil
}
