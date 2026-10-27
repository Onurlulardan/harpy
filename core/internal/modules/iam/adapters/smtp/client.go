package smtp

import (
	"context"
	"fmt"
	"reflect"
	"time"

	gomail "github.com/wneessen/go-mail"
)

type mailSender interface {
	DialAndSendWithContext(context.Context, ...*gomail.Msg) error
}

type Client struct {
	settings Settings
	sender   mailSender
	probe    func(context.Context) error
	now      func() time.Time
}

func Open(settings Settings) (*Client, error) {
	mailClient, err := gomail.NewClient(settings.host, settings.clientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("build IAM SMTP client: %w", ErrContract)
	}
	probe := func(ctx context.Context) error {
		connection, err := mailClient.DialToSMTPClientWithContext(ctx)
		if err != nil {
			return err
		}
		return mailClient.CloseWithSMTPClient(connection)
	}
	return newClient(settings, mailClient, probe, time.Now)
}

func newClient(
	settings Settings,
	sender mailSender,
	probe func(context.Context) error,
	now func() time.Time,
) (*Client, error) {
	if sender == nil || probe == nil || now == nil || typedNil(sender) ||
		settings.host == "" {
		return nil, ErrContract
	}
	return &Client{settings: settings, sender: sender, probe: probe, now: now}, nil
}

func typedNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	return reflected.Kind() == reflect.Pointer && reflected.IsNil()
}

func (client *Client) Probe(ctx context.Context) error {
	if client == nil || client.probe == nil || ctx == nil || ctx.Err() != nil {
		return ErrContract
	}
	if err := client.probe(ctx); err != nil {
		return fmt.Errorf("IAM SMTP dependency preflight: %w", err)
	}
	return nil
}
