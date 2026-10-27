package systemresources

import publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"

type DeleteResult struct {
	deleted      publicresource.DeleteResult
	presentation Presentation
	validated    bool
}

func NewDeleteResult(
	deleted publicresource.DeleteResult,
	presentation Presentation,
) (DeleteResult, error) {
	result := DeleteResult{
		deleted: deleted, presentation: presentation, validated: true,
	}
	if result.Validate() != nil {
		return DeleteResult{}, ErrContract
	}
	return result, nil
}

func (result DeleteResult) Validate() error {
	if !result.validated || result.deleted.Validate() != nil ||
		result.presentation.Validate() != nil {
		return ErrContract
	}
	deletedCatalog, err := result.deleted.Catalog()
	presentationCatalog, presentationErr := result.presentation.Catalog()
	if err != nil || presentationErr != nil {
		return ErrContract
	}
	equal, err := deletedCatalog.ExactEqual(presentationCatalog)
	if err != nil || !equal {
		return ErrContract
	}
	return nil
}

func (result DeleteResult) IAM() (publicresource.DeleteResult, error) {
	if result.Validate() != nil {
		return publicresource.DeleteResult{}, ErrContract
	}
	return result.deleted, nil
}

func (result DeleteResult) Presentation() (Presentation, error) {
	if result.Validate() != nil {
		return Presentation{}, ErrContract
	}
	return result.presentation, nil
}
