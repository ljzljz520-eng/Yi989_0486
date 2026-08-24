package model

import (
	"fmt"
	"strings"
	"time"
)

func FormatMoney(cents int64, currency string) string {
	whole, fraction := cents/100, cents%100
	if fraction < 0 {
		fraction = -fraction
	}
	if currency == "" {
		currency = "CNY"
	}
	return fmt.Sprintf("%s %d.%02d", strings.ToUpper(currency), whole, fraction)
}

func FormatWindow(start, end time.Time) string {
	if start.IsZero() || end.IsZero() {
		return "unscheduled"
	}
	return start.Format("2006-01-02 15:04") + " - " + end.Format("2006-01-02 15:04")
}

func DescribeBatch(batch Batch, now time.Time) string {
	state := StatusLabel(batch.Status)
	availability := "unavailable"
	if batch.IsAvailable(now) {
		availability = "available"
	}
	return fmt.Sprintf("%s (%s): %s; %d/%d remaining; %s", batch.Name, state, availability, batch.Remaining, batch.TotalStock, FormatWindow(batch.StartAt, batch.EndAt))
}

func RecordLabel(record Record) string {
	if record.CustomerID == "" {
		return record.ID
	}
	return fmt.Sprintf("%s for %s", record.ID, record.CustomerID)
}

func ActionLabel(action string) string {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "create":
		return "Created"
	case "edit":
		return "Edited"
	case "delete":
		return "Deleted"
	case "export":
		return "Exported"
	case "claim":
		return "Claimed"
	case "confirm":
		return "Confirmed"
	case "reclaim":
		return "Reclaimed"
	default:
		return "Updated"
	}
}
