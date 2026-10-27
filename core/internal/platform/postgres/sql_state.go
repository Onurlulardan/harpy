package postgres

import "errors"

type sqlStateCarrier interface {
	SQLState() string
}

func ErrorHasSQLState(err error, allowed ...string) bool {
	var carrier sqlStateCarrier
	if !errors.As(err, &carrier) {
		return false
	}
	state := carrier.SQLState()
	for _, candidate := range allowed {
		if state == candidate {
			return true
		}
	}
	return false
}
