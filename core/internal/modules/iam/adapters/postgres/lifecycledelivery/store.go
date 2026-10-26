package lifecycledelivery

import lifecycledeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycledelivery"

type store struct{ database deliveryDatabase }

var _ lifecycledeliveryapp.Store = (*store)(nil)
