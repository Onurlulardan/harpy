package smtp

import (
	"context"
	"crypto/tls"
	"net"
	netsmtp "net/smtp"
	"time"
)

type ipResolver interface {
	LookupIPAddr(context.Context, string) ([]net.IPAddr, error)
}

type contextDialer interface {
	DialContext(context.Context, string, string) (net.Conn, error)
}

type networkExchanger struct {
	resolver ipResolver
	dialer   contextDialer
	now      func() time.Time
}

func newNetworkExchanger() *networkExchanger {
	return &networkExchanger{resolver: net.DefaultResolver,
		dialer: &net.Dialer{Timeout: 15 * time.Second}, now: time.Now}
}

func (exchange *networkExchanger) Exchange(ctx context.Context,
	request exchangeRequest,
) exchangeResult {
	connection, err := exchange.dial(ctx, request.host, request.port)
	if err != nil {
		return exchange.failure("connect", err, false)
	}
	defer connection.Close()
	deadline := exchange.now().Add(30 * time.Second)
	if value, ok := ctx.Deadline(); ok && value.Before(deadline) {
		deadline = value
	}
	_ = connection.SetDeadline(deadline)
	client, err := exchange.smtpClient(ctx, connection, request)
	if err != nil {
		return exchange.failure("tls", err, false)
	}
	defer client.Close()
	if err = client.Auth(netsmtp.PlainAuth("", string(request.username),
		string(request.password), request.host)); err != nil {
		return exchange.failure("auth", err, false)
	}
	if err = client.Mail(request.from); err != nil {
		return exchange.failure("mail_from", err, false)
	}
	if err = client.Rcpt(request.target); err != nil {
		return exchange.failure("recipient", err, false)
	}
	writer, err := client.Data()
	if err != nil {
		return exchange.failure("data_start", err, false)
	}
	if _, err = writer.Write(request.message); err != nil {
		_ = writer.Close()
		return exchange.failure("data_write", err, true)
	}
	if err = writer.Close(); err != nil {
		return exchange.failure("data_commit", err, true)
	}
	_ = client.Quit()
	return exchangeResult{accepted: true, definitive: true,
		code: "SMTP_ACCEPTED", class: "data_commit", statusCode: "250",
		bytesWritten: true, completedAt: exchange.now().UTC()}
}

func (exchange *networkExchanger) smtpClient(ctx context.Context, connection net.Conn,
	request exchangeRequest,
) (*netsmtp.Client, error) {
	configuration := &tls.Config{ServerName: request.host, MinVersion: tls.VersionTLS12}
	if request.security == "implicit_tls" {
		secured := tls.Client(connection, configuration)
		if err := secured.HandshakeContext(ctx); err != nil {
			return nil, err
		}
		return netsmtp.NewClient(secured, request.host)
	}
	client, err := netsmtp.NewClient(connection, request.host)
	if err != nil {
		return nil, err
	}
	if err = client.StartTLS(configuration); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}

var _ exchanger = (*networkExchanger)(nil)
