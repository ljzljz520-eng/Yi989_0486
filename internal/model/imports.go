package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ImportIssue struct {
	Row     int    `json:"row"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

type BatchImport struct {
	Batches []Batch       `json:"batches"`
	Issues  []ImportIssue `json:"issues"`
}

func ParseBatchRow(row int, columns []string, now time.Time) (Batch, []ImportIssue) {
	issues := make([]ImportIssue, 0)
	get := func(index int, field string) string {
		if index >= len(columns) {
			issues = append(issues, ImportIssue{Row: row, Field: field, Message: "missing value"})
			return ""
		}
		return strings.TrimSpace(columns[index])
	}
	batch := Batch{ID: get(0, "id"), Name: get(1, "name"), Description: get(2, "description"), Currency: get(4, "currency"), StartAt: now, EndAt: now.Add(24 * time.Hour), Status: StatusDraft}
	amount := get(3, "amount_cents")
	if amount != "" {
		parsed, err := strconv.ParseInt(amount, 10, 64)
		if err != nil {
			issues = append(issues, ImportIssue{Row: row, Field: "amount_cents", Message: "must be an integer"})
		} else {
			batch.AmountCents = parsed
		}
	}
	stock := get(5, "total_stock")
	if stock != "" {
		parsed, err := strconv.Atoi(stock)
		if err != nil {
			issues = append(issues, ImportIssue{Row: row, Field: "total_stock", Message: "must be an integer"})
		} else {
			batch.TotalStock, batch.Remaining = parsed, parsed
		}
	}
	start := get(6, "start_at")
	if start != "" {
		parsed, err := time.Parse(time.RFC3339, start)
		if err != nil {
			issues = append(issues, ImportIssue{Row: row, Field: "start_at", Message: "must be RFC3339"})
		} else {
			batch.StartAt = parsed
		}
	}
	end := get(7, "end_at")
	if end != "" {
		parsed, err := time.Parse(time.RFC3339, end)
		if err != nil {
			issues = append(issues, ImportIssue{Row: row, Field: "end_at", Message: "must be RFC3339"})
		} else {
			batch.EndAt = parsed
		}
	}
	status := get(8, "status")
	if status != "" {
		batch.Status = BatchStatus(status)
	}
	if err := batch.Validate(now); err != nil {
		issues = append(issues, ImportIssue{Row: row, Field: "batch", Message: err.Error()})
	}
	return batch, issues
}

func ValidateImport(input BatchImport) error {
	if len(input.Batches) == 0 && len(input.Issues) == 0 {
		return errors.New("import has no rows")
	}
	if len(input.Issues) > 0 {
		return fmt.Errorf("import has %d issues", len(input.Issues))
	}
	seen := make(map[string]struct{}, len(input.Batches))
	for _, batch := range input.Batches {
		if _, ok := seen[batch.ID]; ok {
			return fmt.Errorf("duplicate batch id %s", batch.ID)
		}
		seen[batch.ID] = struct{}{}
	}
	return nil
}
