package billing

import (
	"fmt"

	offeringquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/offering"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func globalHTTPQueryKeys(keyring config.CursorKeyring) (offeringquery.CursorKeys, error) {
	current, err := keyring.Current()
	if err != nil {
		return offeringquery.CursorKeys{}, fmt.Errorf("load Billing REST Query current key: %w", err)
	}
	previous, err := keyring.Previous()
	if err != nil {
		return offeringquery.CursorKeys{}, fmt.Errorf("load Billing REST Query previous key: %w", err)
	}
	currentID, currentIDErr := current.KID()
	currentKey, currentKeyErr := current.AES256Key()
	previousID, previousIDErr := previous.KID()
	previousKey, previousKeyErr := previous.AES256Key()
	if currentIDErr != nil || currentKeyErr != nil ||
		previousIDErr != nil || previousKeyErr != nil {
		clear(currentKey[:])
		clear(previousKey[:])
		return offeringquery.CursorKeys{}, fmt.Errorf("load Billing REST Query cursor keys")
	}
	keys, err := offeringquery.NewCursorKeys(
		currentID, currentKey, previousID, previousKey,
	)
	clear(currentKey[:])
	clear(previousKey[:])
	return keys, err
}
