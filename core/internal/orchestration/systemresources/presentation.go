package systemresources

import (
	"crypto/subtle"

	formpresentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

type Presentation struct {
	catalog        publicresource.FieldCatalog
	document       formpresentation.Document
	catalogSHA256  [32]byte
	documentSHA256 [32]byte
	validated      bool
}

func NewPresentation(
	catalog publicresource.FieldCatalog,
	document formpresentation.Document,
) (Presentation, error) {
	fieldCatalog, err := catalog.Presentation()
	if err != nil {
		return Presentation{}, ErrContract
	}
	catalogSHA, err := fieldCatalog.CanonicalSHA256()
	if err != nil {
		return Presentation{}, ErrContract
	}
	_, documentSHA, err := formpresentation.EncodeDocument(document, fieldCatalog)
	if err != nil {
		return Presentation{}, ErrUnavailable
	}
	result := Presentation{
		catalog: catalog, document: formpresentation.CloneDocument(document),
		catalogSHA256: catalogSHA, documentSHA256: documentSHA, validated: true,
	}
	if result.Validate() != nil {
		return Presentation{}, ErrUnavailable
	}
	return result, nil
}

func (value Presentation) Validate() error {
	fieldCatalog, err := value.catalog.Presentation()
	if !value.validated || err != nil {
		return ErrContract
	}
	catalogSHA, err := fieldCatalog.CanonicalSHA256()
	if err != nil || subtle.ConstantTimeCompare(
		catalogSHA[:], value.catalogSHA256[:],
	) != 1 {
		return ErrContract
	}
	_, documentSHA, err := formpresentation.EncodeDocument(value.document, fieldCatalog)
	if err != nil || subtle.ConstantTimeCompare(
		documentSHA[:], value.documentSHA256[:],
	) != 1 {
		return ErrContract
	}
	return nil
}

func (value Presentation) Catalog() (publicresource.FieldCatalog, error) {
	if value.Validate() != nil {
		return publicresource.FieldCatalog{}, ErrContract
	}
	return value.catalog, nil
}

func (value Presentation) Document() (formpresentation.Document, error) {
	if value.Validate() != nil {
		return formpresentation.Document{}, ErrContract
	}
	return formpresentation.CloneDocument(value.document), nil
}

func (value Presentation) Evidence() ([32]byte, [32]byte, error) {
	if value.Validate() != nil {
		return [32]byte{}, [32]byte{}, ErrContract
	}
	return value.catalogSHA256, value.documentSHA256, nil
}
