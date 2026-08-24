package model

import (
	"fmt"
	"strings"
	"time"
)

func ValidateCurrency(value string) error {
	value = strings.ToUpper(strings.TrimSpace(value))
	if len(value) != 3 {
		return fmt.Errorf("currency must be three letters")
	}
	for _, runeValue := range value {
		if runeValue < 'A' || runeValue > 'Z' {
			return fmt.Errorf("currency contains invalid characters")
		}
	}
	return nil
}

func ValidateCustomerID(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("customer id is required")
	}
	if len(value) > 128 {
		return fmt.Errorf("customer id is too long")
	}
	return nil
}

func ValidateDateRange(start, end time.Time, now time.Time) error {
	if start.IsZero() || end.IsZero() {
		return fmt.Errorf("date range is required")
	}
	if !end.After(start) {
		return fmt.Errorf("end must follow start")
	}
	if end.Before(now.Add(-365 * 24 * time.Hour)) {
		return fmt.Errorf("date range is too old")
	}
	return nil
}

func NormalizeCurrency(value string) string { return strings.ToUpper(strings.TrimSpace(value)) }

func NormalizeName(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func IsTerminal(status RecordState) bool {
	return status == RecordReclaimed || status == RecordCancelled
}

func CanTransition(from, to RecordState) bool {
	switch from {
	case RecordReserved:
		return to == RecordConfirmed || to == RecordReclaimed || to == RecordCancelled
	case RecordConfirmed:
		return to == RecordReclaimed
	case RecordReclaimed, RecordCancelled:
		return false
	default:
		return false
	}
}
