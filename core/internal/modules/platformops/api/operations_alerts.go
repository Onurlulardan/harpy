package api

import (
	"context"
	"time"
)

type AlertCondition string

const (
	AlertUnhealthy AlertCondition = "unhealthy"
	AlertDegraded  AlertCondition = "degraded"
	AlertAnyChange AlertCondition = "any_change"
)

type AlertChannel struct {
	ID, Type string
	Config   map[string]any
	Enabled  bool
}

type AlertRule struct {
	ID, EnvironmentID, Name, Description, ServiceName string
	Enabled                                           bool
	Condition                                         AlertCondition
	CooldownMinutes                                   int
	Channels                                          []AlertChannel
	CreatedAt, UpdatedAt                              time.Time
}

type AlertRuleInput struct {
	EnvironmentID, Name, Description, ServiceName string
	Enabled                                       bool
	Condition                                     AlertCondition
	CooldownMinutes                               int
	Channels                                      []AlertChannel
}

type AlertEvent struct {
	ID, EnvironmentID, RuleID, ServiceName string
	PreviousStatus, CurrentStatus          HealthStatus
	ChannelsUsed                           []string
	Message                                string
	ResolvedAt                             *time.Time
	CreatedAt                              time.Time
}

type AlertEventFilter struct {
	Page, Limit         int
	ServiceName, Status string
}

type AlertEventPage struct {
	Items                   []AlertEvent
	Total                   int64
	Page, Limit, TotalPages int
}

type AlertOperations interface {
	ListAlertRules(context.Context, string) ([]AlertRule, error)
	GetAlertRule(context.Context, string, string) (AlertRule, error)
	CreateAlertRule(context.Context, AlertRuleInput) (AlertRule, error)
	UpdateAlertRule(context.Context, string, AlertRuleInput) (AlertRule, error)
	DeleteAlertRule(context.Context, string, string) error
	RecordAlertTest(context.Context, string, string, string) (AlertEvent, error)
	ListAlertEvents(context.Context, string, AlertEventFilter) (AlertEventPage, error)
	GetAlertEvent(context.Context, string, string) (AlertEvent, error)
}
