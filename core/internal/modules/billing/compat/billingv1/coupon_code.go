package billingv1

import (
	"crypto/rand"
	"fmt"
	"io"
	"strings"
)

const couponCodeAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func secureCouponCode() (string, error) {
	const unbiasedLimit = byte(252)
	code := make([]byte, 8)
	buffer := []byte{0}
	for position := 0; position < len(code); {
		if _, err := io.ReadFull(rand.Reader, buffer); err != nil {
			return "", fmt.Errorf("generate coupon code: %w", err)
		}
		if buffer[0] >= unbiasedLimit {
			continue
		}
		code[position] = couponCodeAlphabet[int(buffer[0])%len(couponCodeAlphabet)]
		position++
	}
	return string(code[:4]) + "-" + string(code[4:]), nil
}

func (handler *Handler) couponCode(input *string) (string, error) {
	if input != nil && *input != "" {
		return strings.ToUpper(*input), nil
	}
	code, err := handler.couponCodeSource()
	if err != nil {
		return "", ErrRequestInvalid
	}
	return code, nil
}
