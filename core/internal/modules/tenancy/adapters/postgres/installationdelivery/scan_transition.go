package installationdelivery

import installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"

func scanTransition(row Row) (installationdelivery.Transition, error) {
	var result installationdelivery.Transition
	var state string
	if err := row.Scan(&state, &result.Revision, &result.DatabaseNow); err != nil {
		return installationdelivery.Transition{}, installationdelivery.ErrUnavailable
	}
	result.DatabaseNow = result.DatabaseNow.UTC()
	result.State = installationdelivery.State(state)
	return result, nil
}
