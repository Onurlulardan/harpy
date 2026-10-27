package postgres

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/platform/actorbridge"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func classifyActorBridgeError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return actorbridge.NewFault(actorbridge.FaultUnavailable)
	}
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		return actorbridge.NewFault(actorbridge.FaultUnavailable)
	}
	switch postgresError.Code {
	case "22023":
		return actorbridge.NewFault(actorbridge.FaultInvalid)
	case "42501":
		return actorbridge.NewFault(actorbridge.FaultUnauthorized)
	case "P5601", "P5602":
		return actorbridge.NewFault(actorbridge.FaultAuthority)
	case "P5603", "P5607":
		return actorbridge.NewFault(actorbridge.FaultRetired)
	case "P5604":
		return actorbridge.NewFault(actorbridge.FaultReplay)
	case "P5605":
		return actorbridge.NewFault(actorbridge.FaultExpired)
	case "P5606":
		return actorbridge.NewFault(actorbridge.FaultUnreconciled)
	default:
		return actorbridge.NewFault(actorbridge.FaultUnavailable)
	}
}
