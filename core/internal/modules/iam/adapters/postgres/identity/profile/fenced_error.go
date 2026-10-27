package profilepostgres

import (
	"errors"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

const profileRecoveryRequiredSQLState = "P5741"

type profileSQLStateError interface {
	SQLState() string
}

func profileContractError(err error) error {
	var state profileSQLStateError
	if errors.As(err, &state) &&
		state.SQLState() == profileRecoveryRequiredSQLState {
		return profileapp.ErrRecoveryRequired
	}
	return identitydomain.ErrStorage
}
