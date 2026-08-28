package service

import (
	"context"
	"fmt"
	"time"

	"couponbatch/internal/model"
	"couponbatch/internal/permission"
)

type BulkResult struct {
	Requested int      `json:"requested"`
	Changed   int      `json:"changed"`
	Failed    int      `json:"failed"`
	Errors    []string `json:"errors,omitempty"`
}

func (svc *BatchService) BulkPause(ctx context.Context, actor model.Profile, ids []string) BulkResult {
	result := BulkResult{Requested: len(ids), Errors: make([]string, 0)}
	if !permission.Allows(actor, permission.Edit) {
		result.Failed = len(ids)
		result.Errors = append(result.Errors, model.ErrUnauthorized.Error())
		return result
	}
	for _, id := range ids {
		if _, err := svc.PauseBatch(ctx, actor, id); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", id, err))
		} else {
			result.Changed++
		}
	}
	return result
}

func (svc *BatchService) BulkArchive(ctx context.Context, actor model.Profile, ids []string) BulkResult {
	result := BulkResult{Requested: len(ids), Errors: make([]string, 0)}
	for _, id := range ids {
		if _, err := svc.Archive(ctx, actor, id); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", id, err))
		} else {
			result.Changed++
		}
	}
	return result
}

func (svc *BatchService) Rebalance(ctx context.Context, actor model.Profile, id string, stock int) (model.Batch, error) {
	if stock <= 0 {
		return model.Batch{}, fmt.Errorf("stock must be positive")
	}
	batch, err := svc.Store.GetBatch(id)
	if err != nil {
		return model.Batch{}, err
	}
	if batch.Issued > stock {
		return model.Batch{}, fmt.Errorf("stock below issued")
	}
	batch.TotalStock = stock
	batch.Remaining = stock - batch.Issued
	return svc.UpdateBatch(ctx, actor, batch)
}

func (svc *BatchService) ActivateDue(ctx context.Context, actor model.Profile) (int, error) {
	if !permission.Allows(actor, permission.Edit) {
		return 0, model.ErrUnauthorized
	}
	items, err := svc.Store.ListBatches()
	if err != nil {
		return 0, err
	}
	changed := 0
	now := svc.now()
	for _, item := range items {
		if item.Status == model.StatusDraft && !item.StartAt.After(now) && item.EndAt.After(now) {
			item.Status = model.StatusActive
			if _, err := svc.UpdateBatch(ctx, actor, item); err != nil {
				return changed, err
			}
			changed++
		}
	}
	return changed, nil
}

func (svc *ClaimService) ReclaimExpired(ctx context.Context, actor model.Profile, before time.Time) (int, error) {
	if !permission.Allows(actor, permission.Claim) {
		return 0, model.ErrUnauthorized
	}
	records, err := svc.Store.ListRecords()
	if err != nil {
		return 0, err
	}
	changed := 0
	for _, record := range records {
		if record.State == model.RecordConfirmed && record.ConfirmedAt != nil && record.ConfirmedAt.Before(before) {
			if _, err := svc.Reclaim(ctx, actor, record.ID, "expired claim"); err != nil {
				return changed, err
			}
			changed++
		}
	}
	return changed, nil
}
