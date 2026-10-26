package transaction

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const providerWebhookReceiveOperation = "billing.provider.webhook.receive"

var providerKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9._-]{0,62}$`)
var providerAccountPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,254}$`)

type ProviderLookup struct {
	provider          string
	accountID         string
	objectKind        string
	objectID          string
	eventID           string
	authenticationSHA [32]byte
}

func NewProviderLookup(
	provider, accountID, objectKind, objectID, eventID string,
	authenticationSHA256 [32]byte,
) (ProviderLookup, error) {
	lookup := ProviderLookup{
		provider: provider, accountID: accountID, objectKind: objectKind,
		objectID: objectID, eventID: eventID, authenticationSHA: authenticationSHA256,
	}
	if !lookup.valid() {
		return ProviderLookup{}, fmt.Errorf("provider lookup is invalid")
	}
	return lookup, nil
}

func (lookup ProviderLookup) OperationID() string { return providerWebhookReceiveOperation }
func (lookup ProviderLookup) Provider() string    { return lookup.provider }
func (lookup ProviderLookup) AccountID() string   { return lookup.accountID }
func (lookup ProviderLookup) ObjectKind() string  { return lookup.objectKind }
func (lookup ProviderLookup) ObjectID() string    { return lookup.objectID }
func (lookup ProviderLookup) EventID() string     { return lookup.eventID }
func (lookup ProviderLookup) AuthenticationSHA256() [32]byte {
	return lookup.authenticationSHA
}

func (lookup ProviderLookup) valid() bool {
	return providerKeyPattern.MatchString(lookup.provider) &&
		providerAccountPattern.MatchString(lookup.accountID) &&
		validProviderObjectKind(lookup.objectKind) &&
		validProviderIdentifier(lookup.objectID, 512) &&
		validProviderIdentifier(lookup.eventID, 512) &&
		lookup.authenticationSHA != [32]byte{}
}

func validProviderObjectKind(value string) bool {
	return value == "customer" || value == "subscription" ||
		value == "invoice" || value == "payment"
}

func validProviderIdentifier(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && utf8.ValidString(value) &&
		strings.TrimSpace(value) == value && !strings.ContainsAny(value, "\x00\r\n")
}
