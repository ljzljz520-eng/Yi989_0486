package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"couponbatch/internal/model"
)

type batchUpdateRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	AmountCents int64             `json:"amount_cents"`
	Currency    string            `json:"currency"`
	TotalStock  int               `json:"total_stock"`
	StartAt     string            `json:"start_at"`
	EndAt       string            `json:"end_at"`
	Status      model.BatchStatus `json:"status"`
	Version     int               `json:"version"`
	Tags        []string          `json:"tags"`
}

func parseUpdate(data batchUpdateRequest) (model.Batch, error) {
	batch := model.Batch{Name: strings.TrimSpace(data.Name), Description: data.Description, AmountCents: data.AmountCents, Currency: data.Currency, TotalStock: data.TotalStock, Status: data.Status, Version: data.Version, Tags: data.Tags}
	if data.StartAt != "" {
		parsed, err := parseTime(data.StartAt)
		if err != nil {
			return model.Batch{}, err
		}
		batch.StartAt = parsed
	}
	if data.EndAt != "" {
		parsed, err := parseTime(data.EndAt)
		if err != nil {
			return model.Batch{}, err
		}
		batch.EndAt = parsed
	}
	return batch, nil
}

func parseTime(value string) (time.Time, error) { return time.Parse(time.RFC3339, value) }

func writeMethodError(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func decodeUpdate(r *http.Request) (model.Batch, error) {
	var request batchUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return model.Batch{}, model.ErrInvalidRequest
	}
	return parseUpdate(request)
}
