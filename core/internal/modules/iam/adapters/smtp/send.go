package smtp

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"

	gomail "github.com/wneessen/go-mail"
)

func (client *Client) Send(ctx context.Context, value Message) Result {
	if client == nil || client.sender == nil || ctx == nil || ctx.Err() != nil {
		return Result{Kind: ResultUnknown, Code: "SMTP_OUTCOME_UNKNOWN"}
	}
	message, err := client.buildMessage(value)
	if err != nil {
		return client.result(ResultPermanent, "SMTP_MESSAGE_INVALID", value.MessageID, 0, "")
	}
	defer message.UnsetAllParts()
	bounded, cancel := context.WithTimeout(ctx, client.settings.timeout)
	defer cancel()
	err = client.sender.DialAndSendWithContext(bounded, message)
	if message.IsDelivered() {
		return client.result(
			ResultAccepted, "SMTP_DATA_ACCEPTED", value.MessageID, 250,
			message.ServerResponse(),
		)
	}
	kind, code, status := classifySendError(err)
	return client.result(kind, code, value.MessageID, status, "")
}

func classifySendError(err error) (ResultKind, string, int) {
	if err == nil {
		return ResultUnknown, "SMTP_OUTCOME_UNKNOWN", 0
	}
	var sendError *gomail.SendError
	if !errors.As(err, &sendError) || sendError == nil {
		return ResultTransient, "SMTP_UNAVAILABLE", 0
	}
	code := sendError.ErrorCode()
	switch sendError.Reason {
	case gomail.ErrSMTPMailFrom:
		return explicitSMTPFailure(sendError, "SMTP_SENDER_REJECTED")
	case gomail.ErrSMTPRcptTo:
		return explicitSMTPFailure(sendError, "SMTP_RECIPIENT_REJECTED")
	case gomail.ErrSMTPData:
		return explicitSMTPFailure(sendError, "SMTP_DATA_REJECTED")
	case gomail.ErrSMTPDataClose:
		if code >= 400 && code <= 599 {
			return explicitSMTPFailure(sendError, "SMTP_MESSAGE_REJECTED")
		}
		return ResultUnknown, "SMTP_OUTCOME_UNKNOWN", code
	case gomail.ErrGetSender, gomail.ErrGetRcpts, gomail.ErrNoUnencoded:
		return ResultPermanent, "SMTP_MESSAGE_INVALID", code
	case gomail.ErrConnCheck:
		return ResultTransient, "SMTP_UNAVAILABLE", code
	default:
		return ResultUnknown, "SMTP_OUTCOME_UNKNOWN", code
	}
}

func explicitSMTPFailure(value *gomail.SendError, code string) (ResultKind, string, int) {
	if value.IsTemp() || value.ErrorCode() >= 400 && value.ErrorCode() <= 499 {
		return ResultTransient, "SMTP_TEMPORARY_FAILURE", value.ErrorCode()
	}
	return ResultPermanent, code, value.ErrorCode()
}

func (client *Client) result(
	kind ResultKind,
	code string,
	reference string,
	status int,
	response string,
) Result {
	hash := sha256.New()
	for _, part := range []string{"iam.smtp.result.v1", code, reference, response} {
		var length [4]byte
		binary.BigEndian.PutUint32(length[:], uint32(len(part)))
		_, _ = hash.Write(length[:])
		_, _ = hash.Write([]byte(part))
	}
	var encodedStatus [4]byte
	binary.BigEndian.PutUint32(encodedStatus[:], uint32(status))
	_, _ = hash.Write(encodedStatus[:])
	var digest [32]byte
	copy(digest[:], hash.Sum(nil))
	return Result{
		Kind: kind, Code: code, ProviderReference: reference,
		EvidenceSHA256: digest, CompletedAt: client.now().UTC(),
	}
}
