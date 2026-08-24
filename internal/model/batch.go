package model

import (
	"errors"
	"strings"
	"time"
)

type BatchStatus string

const (
	StatusDraft    BatchStatus = "draft"
	StatusActive   BatchStatus = "active"
	StatusPaused   BatchStatus = "paused"
	StatusArchived BatchStatus = "archived"
)

type Batch struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	AmountCents int64       `json:"amount_cents"`
	Currency    string      `json:"currency"`
	TotalStock  int         `json:"total_stock"`
	Remaining   int         `json:"remaining"`
	Issued      int         `json:"issued"`
	StartAt     time.Time   `json:"start_at"`
	EndAt       time.Time   `json:"end_at"`
	Status      BatchStatus `json:"status"`
	CreatedBy   string      `json:"created_by"`
	UpdatedBy   string      `json:"updated_by"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Version     int         `json:"version"`
	Tags        []string    `json:"tags"`
}

type BatchFilter struct {
	Query       string
	Status      BatchStatus
	MinAmount   int64
	MaxAmount   int64
	Available   bool
	StartsAfter time.Time
	EndsBefore  time.Time
	Tag         string
}

type BatchPage struct {
	Items      []Batch `json:"items"`
	Total      int     `json:"total"`
	NextCursor string  `json:"next_cursor"`
}

func (b Batch) Validate(now time.Time) error {
	if strings.TrimSpace(b.ID) == "" {
		return errors.New("batch id is required")
	}
	if strings.TrimSpace(b.Name) == "" {
		return errors.New("batch name is required")
	}
	if b.AmountCents <= 0 {
		return errors.New("amount must be positive")
	}
	if b.TotalStock <= 0 {
		return errors.New("stock must be positive")
	}
	if b.Remaining < 0 || b.Remaining > b.TotalStock {
		return errors.New("remaining stock is invalid")
	}
	if b.StartAt.IsZero() || b.EndAt.IsZero() || !b.EndAt.After(b.StartAt) {
		return errors.New("validity window is invalid")
	}
	if b.EndAt.Before(now) && b.Status == StatusActive {
		return errors.New("expired batch cannot be active")
	}
	if b.Currency == "" {
		return errors.New("currency is required")
	}
	if b.Status != StatusDraft && b.Status != StatusActive && b.Status != StatusPaused && b.Status != StatusArchived {
		return errors.New("unknown batch status")
	}
	return nil
}

func (b Batch) IsAvailable(now time.Time) bool {
	return b.Status == StatusActive && b.Remaining > 0 && !now.Before(b.StartAt) && now.Before(b.EndAt)
}

func (b Batch) Clone() Batch {
	copyBatch := b
	copyBatch.Tags = append([]string(nil), b.Tags...)
	return copyBatch
}

func (f BatchFilter) Match(b Batch, now time.Time) bool {
	if f.Status != "" && b.Status != f.Status {
		return false
	}
	if f.Query != "" {
		q := strings.ToLower(strings.TrimSpace(f.Query))
		if !strings.Contains(strings.ToLower(b.Name), q) && !strings.Contains(strings.ToLower(b.Description), q) {
			return false
		}
	}
	if f.MinAmount > 0 && b.AmountCents < f.MinAmount {
		return false
	}
	if f.MaxAmount > 0 && b.AmountCents > f.MaxAmount {
		return false
	}
	if f.Available && !b.IsAvailable(now) {
		return false
	}
	if !f.StartsAfter.IsZero() && b.StartAt.Before(f.StartsAfter) {
		return false
	}
	if !f.EndsBefore.IsZero() && b.EndAt.After(f.EndsBefore) {
		return false
	}
	if f.Tag != "" {
		found := false
		for _, tag := range b.Tags {
			if tag == f.Tag {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
