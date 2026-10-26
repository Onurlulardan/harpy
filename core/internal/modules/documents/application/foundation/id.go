package foundation

import "github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"

type ID = domain.ID

var ErrInvalidID = domain.ErrInvalidID

func ParseID(value string) (ID, error) { return domain.ParseID(value) }
