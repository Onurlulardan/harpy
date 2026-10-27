package api

import (
	"context"
	"time"
)

type HealthStatus string

const (
	HealthHealthy   HealthStatus = "healthy"
	HealthDegraded  HealthStatus = "degraded"
	HealthUnhealthy HealthStatus = "unhealthy"
	HealthUnknown   HealthStatus = "unknown"
)

type HealthComponent struct {
	Name, URL, Error string
	Status           HealthStatus
	ResponseMillis   int64
	LastChecked      time.Time
}

type SystemHealth struct {
	Status         HealthStatus
	ObservedAt     time.Time
	Services       []HealthComponent
	Infrastructure []HealthComponent
}

type HealthReader interface {
	ReadSystemHealth(context.Context) (SystemHealth, error)
}
