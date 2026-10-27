package systemresource

import identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"

type BulkDeleteReceiptCommand = identitydomain.BulkDeleteReceiptCommand

const MaximumBulkDeleteItems = identitydomain.MaximumBulkDeleteItems

var ErrBulkDeleteConflict = identitydomain.ErrBulkDeleteConflict
