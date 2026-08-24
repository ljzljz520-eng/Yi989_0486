package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"couponbatch/internal/model"
	"couponbatch/internal/service"
	"couponbatch/internal/store"
)

type Engine struct {
	Batches *service.BatchService
	Claims  *service.ClaimService
	Store   *store.Store
}

func New(s *store.Store) *Engine {
	return &Engine{Store: s, Batches: service.NewBatchService(s), Claims: service.NewClaimService(s)}
}

func (e *Engine) Accept(ctx context.Context, actor model.Profile, batch model.Batch) (model.Batch, error) {
	created, err := e.Batches.CreateBatch(ctx, actor, batch)
	if err != nil {
		return model.Batch{}, err
	}
	if created.Status == model.StatusDraft {
		created.Status = model.StatusActive
		created, err = e.Batches.UpdateBatch(ctx, actor, created)
	}
	return created, err
}

func (e *Engine) Publish(ctx context.Context, actor model.Profile, batchID, customerID string) (model.Record, error) {
	record, err := e.Claims.Claim(ctx, actor, batchID, customerID)
	if err != nil {
		return model.Record{}, err
	}
	if _, err = e.Claims.Confirm(ctx, actor, record.ID); err != nil {
		return model.Record{}, err
	}
	return e.Claims.GetRecord(ctx, actor, record.ID)
}

func (e *Engine) Recover(ctx context.Context, actor model.Profile, recordID string) (model.Record, error) {
	return e.Claims.Reclaim(ctx, actor, recordID, "operator recovery")
}

func (e *Engine) RunAcceptPublishRecover(ctx context.Context, actor model.Profile, batch model.Batch, customerID string) (model.Record, error) {
	created, err := e.Accept(ctx, actor, batch)
	if err != nil {
		return model.Record{}, err
	}
	record, err := e.Publish(ctx, actor, created.ID, customerID)
	if err != nil {
		return model.Record{}, err
	}
	return e.Recover(ctx, actor, record.ID)
}

func (e *Engine) ConfirmThenReclaim(ctx context.Context, actor model.Profile, recordID string) (model.Record, error) {
	confirmed, confirmErr := e.Claims.Confirm(ctx, actor, recordID)
	if confirmErr != nil {
		return confirmed, nil
	}
	return e.Claims.Reclaim(ctx, actor, recordID, "confirm then reclaim")
}

func (e *Engine) Reopen(path string) (*Engine, error) {
	if e.Store != nil {
		if err := e.Store.Close(); err != nil {
			return nil, err
		}
	}
	opened, err := store.Open(path)
	if err != nil {
		return nil, err
	}
	return New(opened), nil
}

func BuildSample(now time.Time) model.Batch {
	return model.Batch{ID: "sample", Name: "Developer coupon batch", Description: "Storefront claimable sample", AmountCents: 1000, Currency: "CNY", TotalStock: 10, Remaining: 10, StartAt: now.Add(-time.Hour), EndAt: now.Add(24 * time.Hour), Status: model.StatusActive, Tags: []string{"sample", "developer"}}
}

func MustActive(batch model.Batch) error {
	if batch.Status != model.StatusActive {
		return fmt.Errorf("batch %s is not active", batch.ID)
	}
	if batch.Remaining <= 0 {
		return errors.New("batch is exhausted")
	}
	return nil
}
