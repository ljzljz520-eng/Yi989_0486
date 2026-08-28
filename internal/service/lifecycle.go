package service

import (
	"context"
	"errors"
	"time"

	"couponbatch/internal/model"
	"couponbatch/internal/permission"
)

func (svc *BatchService) Archive(ctx context.Context, actor model.Profile, id string) (model.Batch, error) {
	if !permission.Allows(actor, permission.Edit) {
		return model.Batch{}, model.ErrUnauthorized
	}
	batch, err := svc.Store.GetBatch(id)
	if err != nil {
		return model.Batch{}, err
	}
	if batch.Status == model.StatusArchived {
		return model.Batch{}, model.ErrInvalidState
	}
	batch.Status = model.StatusArchived
	return svc.UpdateBatch(ctx, actor, batch)
}

func (svc *BatchService) Resume(ctx context.Context, actor model.Profile, id string) (model.Batch, error) {
	if !permission.Allows(actor, permission.Edit) {
		return model.Batch{}, model.ErrUnauthorized
	}
	batch, err := svc.Store.GetBatch(id)
	if err != nil {
		return model.Batch{}, err
	}
	if batch.Status != model.StatusPaused {
		return model.Batch{}, errors.New("only paused batch can resume")
	}
	batch.Status = model.StatusActive
	return svc.UpdateBatch(ctx, actor, batch)
}

func (svc *BatchService) Expire(ctx context.Context, actor model.Profile) (int, error) {
	if !permission.Allows(actor, permission.Edit) {
		return 0, model.ErrUnauthorized
	}
	items, err := svc.Store.ListBatches()
	if err != nil {
		return 0, err
	}
	count := 0
	for _, item := range items {
		if item.Status == model.StatusActive && !item.EndAt.After(svc.now()) {
			item.Status = model.StatusPaused
			if _, err := svc.UpdateBatch(ctx, actor, item); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, nil
}

func (svc *ClaimService) ValidateRecord(record model.Record) error {
	if record.ID == "" || record.BatchID == "" || record.CustomerID == "" {
		return model.ErrInvalidRequest
	}
	if record.AmountCents <= 0 {
		return errors.New("record amount must be positive")
	}
	if record.CreatedAt.After(time.Now().Add(time.Minute)) {
		return errors.New("record timestamp is in the future")
	}
	return nil
}

func (svc *ClaimService) History(ctx context.Context, actor model.Profile, entityID string) ([]model.Audit, error) {
	if !permission.Allows(actor, permission.View) {
		return nil, model.ErrUnauthorized
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return svc.Store.ListAudits(entityID)
}
