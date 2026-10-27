package postgres

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func scanPreference(row scanner) (domain.Preference, error) {
	var value domain.Preference
	var recipient string
	var timezone *string
	var start, end *int
	err := row.Scan(&recipient, &value.Category, &value.Channel, &value.Enabled,
		&value.Explicit, &value.VerifiedTarget, &timezone, &start, &end, &value.Revision)
	if err != nil {
		return domain.Preference{}, err
	}
	value.RecipientID, err = decodeID(recipient)
	if timezone != nil || start != nil || end != nil {
		if timezone == nil || start == nil || end == nil {
			return domain.Preference{}, fmt.Errorf("decode notification quiet hours")
		}
		value.QuietHours = &domain.QuietHours{TimeZone: *timezone,
			StartMinute: *start, EndMinute: *end}
	}
	if err != nil || !value.Valid() {
		return domain.Preference{}, fmt.Errorf("decode notification preference")
	}
	return value, nil
}

func quietArguments(value *domain.QuietHours) (any, any, any) {
	if value == nil {
		return nil, nil, nil
	}
	return value.TimeZone, value.StartMinute, value.EndMinute
}
