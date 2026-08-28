package analytics

import (
	"fmt"
	"strings"
	"time"

	"couponbatch/internal/model"
)

type HealthCheck struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail"`
}

func CheckBatches(batches []model.Batch, now time.Time) []HealthCheck {
	checks := make([]HealthCheck, 0, 4)
	checks = append(checks, HealthCheck{Name: "unique_ids", Passed: uniqueBatchIDs(batches), Detail: "batch identifiers are unique"})
	checks = append(checks, HealthCheck{Name: "inventory_bounds", Passed: inventoryBounds(batches), Detail: "remaining stock stays within total stock"})
	checks = append(checks, HealthCheck{Name: "valid_windows", Passed: validWindows(batches), Detail: "all validity windows are ordered"})
	checks = append(checks, HealthCheck{Name: "active_dates", Passed: activeDates(batches, now), Detail: "active batches are not expired"})
	return checks
}

func uniqueBatchIDs(batches []model.Batch) bool {
	seen := make(map[string]struct{})
	for _, batch := range batches {
		if _, ok := seen[batch.ID]; ok {
			return false
		}
		seen[batch.ID] = struct{}{}
	}
	return true
}
func inventoryBounds(batches []model.Batch) bool {
	for _, batch := range batches {
		if batch.Remaining < 0 || batch.Remaining > batch.TotalStock || batch.Issued < 0 || batch.Issued > batch.TotalStock {
			return false
		}
	}
	return true
}
func validWindows(batches []model.Batch) bool {
	for _, batch := range batches {
		if batch.StartAt.IsZero() || batch.EndAt.IsZero() || !batch.EndAt.After(batch.StartAt) {
			return false
		}
	}
	return true
}
func activeDates(batches []model.Batch, now time.Time) bool {
	for _, batch := range batches {
		if batch.Status == model.StatusActive && batch.EndAt.Before(now) {
			return false
		}
	}
	return true
}

func ExplainChecks(checks []HealthCheck) string {
	parts := make([]string, 0, len(checks))
	for _, check := range checks {
		state := "pass"
		if !check.Passed {
			state = "fail"
		}
		parts = append(parts, fmt.Sprintf("%s=%s", check.Name, state))
	}
	return strings.Join(parts, ",")
}
