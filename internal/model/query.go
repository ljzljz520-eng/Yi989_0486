package model

import (
	"sort"
	"strings"
)

func SortBatches(items []Batch, field string, descending bool) {
	sort.SliceStable(items, func(i, j int) bool {
		less := false
		switch strings.ToLower(field) {
		case "name":
			less = items[i].Name < items[j].Name
		case "amount":
			less = items[i].AmountCents < items[j].AmountCents
		case "stock":
			less = items[i].Remaining < items[j].Remaining
		default:
			less = items[i].UpdatedAt.Before(items[j].UpdatedAt)
		}
		if descending {
			return !less && items[i].ID != items[j].ID
		}
		return less
	})
}

func StatusLabel(status BatchStatus) string {
	switch status {
	case StatusDraft:
		return "Draft"
	case StatusActive:
		return "Active"
	case StatusPaused:
		return "Paused"
	case StatusArchived:
		return "Archived"
	default:
		return "Unknown"
	}
}
