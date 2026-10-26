package invoice

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
)

func TestLineInvariant(t *testing.T) {
	line, err := NewLine(
		lineID(1), amount(t, 1000, "USD"), amount(t, 200, "USD"),
		amount(t, 160, "USD"), amount(t, 30, "USD"), amount(t, 10, "USD"),
	)
	if err != nil || line.Total().MinorUnits() != 980 {
		t.Fatalf("total=%d, err=%v", line.Total().MinorUnits(), err)
	}
	verified, err := NewVerifiedLine(
		lineID(1), amount(t, 1000, "USD"), amount(t, 200, "USD"),
		amount(t, 160, "USD"), amount(t, 30, "USD"), amount(t, 10, "USD"),
		amount(t, 980, "USD"),
	)
	if err != nil || verified.Total().MinorUnits() != 980 {
		t.Fatalf("verified total=%d, err=%v", verified.Total().MinorUnits(), err)
	}
}

func TestVerifiedLineRejectsMismatch(t *testing.T) {
	_, err := NewVerifiedLine(
		lineID(1), amount(t, 100, "USD"), amount(t, 0, "USD"),
		amount(t, 0, "USD"), amount(t, 0, "USD"), amount(t, 0, "USD"),
		amount(t, 99, "USD"),
	)
	if err != ErrInvariant {
		t.Fatalf("expected ErrInvariant, got %v", err)
	}
}

func TestLineRejectsInvalidComponents(t *testing.T) {
	zero := amount(t, 0, "USD")
	if _, err := NewLine(lineID(1), amount(t, 10, "USD"),
		amount(t, 11, "USD"), zero, zero, zero); err != ErrLine {
		t.Fatalf("discount: expected ErrLine, got %v", err)
	}
	if _, err := NewLine(lineID(1), amount(t, 10, "USD"),
		zero, amount(t, 1, "EUR"), zero, zero); err != ErrLine {
		t.Fatalf("currency: expected ErrLine, got %v", err)
	}
}

func amount(t *testing.T, units int64, currency string) money.Money {
	t.Helper()
	value, err := money.New(units, currency, money.RegistryVersion)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func lineID(last byte) LineIdentifier {
	var identifier LineIdentifier
	identifier[len(identifier)-1] = last
	return identifier
}
