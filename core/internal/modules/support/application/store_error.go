package application

import supportports "github.com/Onurlulardan/harpy/core/internal/modules/support/ports"

func normalizeStoreError(err error) error { return supportports.NormalizeError(err) }
