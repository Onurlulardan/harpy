package smtp

import (
	"errors"
	"fmt"
	"net/textproto"
)

func (exchange *networkExchanger) failure(stage string, err error,
	bytesWritten bool,
) exchangeResult {
	result := exchangeResult{class: stage, bytesWritten: bytesWritten,
		code: "SMTP_" + stageCode(stage), completedAt: exchange.now().UTC()}
	var protocol *textproto.Error
	if errors.As(err, &protocol) {
		result.definitive = true
		result.statusCode = fmt.Sprintf("%d", protocol.Code)
		result.code = "SMTP_" + result.statusCode
		result.temporary = protocol.Code >= 400 && protocol.Code < 500
	}
	return result
}

func stageCode(stage string) string {
	switch stage {
	case "connect":
		return "CONNECT_FAILED"
	case "tls":
		return "TLS_FAILED"
	case "auth":
		return "AUTH_FAILED"
	case "mail_from":
		return "SENDER_REJECTED"
	case "recipient":
		return "RECIPIENT_REJECTED"
	case "data_start":
		return "DATA_REJECTED"
	default:
		return "TRANSPORT_FAILED"
	}
}
