package worker

import (
	"context"
	"errors"
	"testing"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/deadletterdelivery"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type deadLetterProbeConnector struct {
	session deadletterdelivery.Session
	err     error
}

func (connector deadLetterProbeConnector) Open(
	context.Context,
) (deadletterdelivery.Session, error) {
	return connector.session, connector.err
}

type deadLetterProbeSession struct {
	closeErr error
	closed   bool
}

func (*deadLetterProbeSession) Run(context.Context) error {
	return errors.New("unexpected run")
}

func (session *deadLetterProbeSession) Close() error {
	session.closed = true
	return session.closeErr
}

func TestEmptyConsumerCatalogOmitsDeadLetterWorkloadBeforeDependencies(t *testing.T) {
	catalogs, err := platformasync.NewEventCatalogs(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	workload, enabled, err := buildDeadLetterWorkload(
		nil, nil, config.RabbitMQ{}, catalogs, nil,
	)
	if err != nil || enabled || workload != nil {
		t.Fatalf("empty catalog result = %v, enabled=%t, err=%v", workload, enabled, err)
	}
}

func TestDeadLetterPreflightRequiresOpenSessionAndCleanClose(t *testing.T) {
	if err := probeDeadLetterDelivery(nil, nil); err == nil {
		t.Fatal("incomplete dead-letter preflight passed")
	}
	if err := probeDeadLetterDelivery(
		context.Background(), deadLetterProbeConnector{err: errors.New("offline")},
	); err == nil {
		t.Fatal("offline dead-letter broker passed preflight")
	}
	if err := probeDeadLetterDelivery(
		context.Background(), deadLetterProbeConnector{},
	); err == nil {
		t.Fatal("missing dead-letter session passed preflight")
	}
	session := &deadLetterProbeSession{}
	if err := probeDeadLetterDelivery(
		context.Background(), deadLetterProbeConnector{session: session},
	); err != nil || !session.closed {
		t.Fatalf("valid preflight = closed=%t err=%v", session.closed, err)
	}
	if err := probeDeadLetterDelivery(
		context.Background(), deadLetterProbeConnector{
			session: &deadLetterProbeSession{closeErr: errors.New("close")},
		},
	); err == nil {
		t.Fatal("unclean dead-letter preflight close passed")
	}
}
