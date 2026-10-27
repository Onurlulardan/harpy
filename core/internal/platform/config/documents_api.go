package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/netip"
	"strconv"
	"strings"
)

const documentsClamAVAddressKey = "HARPY_DOCUMENTS_CLAMAV_ADDRESS"

var errInvalidDocumentsAPI = errors.New("Documents API config is invalid")

type DocumentsAPI struct {
	artifacts      DocumentsArtifacts
	clamAVAddress  string
	clamAVOptional bool
	valid          bool
}

func loadDocumentsAPI(values environmentValues,
	environment Environment,
) (DocumentsAPI, error) {
	artifacts, err := loadDocumentsArtifacts(values, environment)
	if err != nil {
		return DocumentsAPI{}, err
	}
	address, optional, err := resolveClamAVAddress(values, environment)
	if err != nil {
		return DocumentsAPI{}, err
	}
	return DocumentsAPI{artifacts: artifacts, clamAVAddress: address,
		clamAVOptional: optional, valid: true}, nil
}

func (settings DocumentsAPI) Validate() error {
	if !settings.valid || settings.artifacts.Validate() != nil ||
		!clamAVAccepted(settings.clamAVAddress, settings.clamAVOptional) {
		return errInvalidDocumentsAPI
	}
	return nil
}

func (settings DocumentsAPI) Artifacts() (DocumentsArtifacts, error) {
	if err := settings.Validate(); err != nil {
		return DocumentsArtifacts{}, err
	}
	return settings.artifacts, nil
}

func (settings DocumentsAPI) ClamAVAddress() (string, error) {
	if err := settings.Validate(); err != nil {
		return "", err
	}
	return settings.clamAVAddress, nil
}

func resolveClamAVAddress(
	values environmentValues,
	environment Environment,
) (string, bool, error) {
	address, found := values.lookup(documentsClamAVAddressKey)
	if !found || address == "" {
		if environment.productionLike() {
			return "", false, configError(documentsClamAVAddressKey,
				"must be a canonical host and port")
		}
		return "", true, nil
	}
	if !validClamAVAddress(address) {
		return "", false, configError(documentsClamAVAddressKey,
			"must be a canonical host and port")
	}
	return address, false, nil
}

func clamAVAccepted(address string, optional bool) bool {
	if optional {
		return address == ""
	}
	return validClamAVAddress(address)
}

func validClamAVAddress(value string) bool {
	if value == "" || value != strings.ToLower(value) ||
		strings.TrimSpace(value) != value {
		return false
	}
	host, rawPort, err := net.SplitHostPort(value)
	if err != nil || host == "" {
		return false
	}
	port, err := strconv.ParseUint(rawPort, 10, 16)
	if err != nil || port == 0 || strconv.FormatUint(port, 10) != rawPort {
		return false
	}
	if address, err := netip.ParseAddr(host); err == nil {
		return address.String() == host
	}
	return validDocumentsHost(host)
}

func validDocumentsHost(value string) bool {
	if len(value) > 253 || strings.Contains(value, "..") {
		return false
	}
	for _, label := range strings.Split(value, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, character := range label {
			if (character < 'a' || character > 'z') &&
				(character < '0' || character > '9') && character != '-' {
				return false
			}
		}
	}
	return true
}

func (DocumentsAPI) String() string   { return "[REDACTED DOCUMENTS API CONFIG]" }
func (DocumentsAPI) GoString() string { return "[REDACTED DOCUMENTS API CONFIG]" }
func (DocumentsAPI) LogValue() slog.Value {
	return slog.StringValue("[REDACTED DOCUMENTS API CONFIG]")
}
func (DocumentsAPI) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED DOCUMENTS API CONFIG]")
}
