package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"couponbatch/internal/model"
	"couponbatch/internal/permission"
	"couponbatch/internal/store"
	"go.etcd.io/bbolt"
)

type Clock func() time.Time

type BatchService struct {
	Store *store.Store
	Now   Clock
}

func NewBatchService(s *store.Store) *BatchService {
	return &BatchService{Store: s, Now: time.Now}
}

func (svc *BatchService) now() time.Time {
	if svc.Now == nil {
		return time.Now()
	}
	return svc.Now()
}

func (svc *BatchService) CreateBatch(ctx context.Context, actor model.Profile, draft model.Batch) (model.Batch, error) {
	if err := ctx.Err(); err != nil {
		return model.Batch{}, err
	}
	if !permission.Allows(actor, permission.Create) {
		return model.Batch{}, model.ErrUnauthorized
	}
	if draft.ID == "" {
		draft.ID = newID("bat")
	}
	draft.Tags = model.NormalizeTags(draft.Tags)
	if draft.Status == "" {
		draft.Status = model.StatusDraft
	}
	if draft.Remaining == 0 {
		draft.Remaining = draft.TotalStock
	}
	draft.CreatedBy, draft.UpdatedBy = actor.ID, actor.ID
	draft.CreatedAt, draft.UpdatedAt = svc.now(), svc.now()
	draft.Version = 1
	if err := draft.Validate(svc.now()); err != nil {
		return model.Batch{}, err
	}
	if _, err := svc.Store.GetBatch(draft.ID); err == nil {
		return model.Batch{}, model.ErrAlreadyExists
	}
	if err := svc.Store.SaveBatch(draft); err != nil {
		return model.Batch{}, err
	}
	_ = svc.recordAudit(draft.ID, "create", actor.ID, "batch created")
	return draft, nil
}

func (svc *BatchService) GetBatch(ctx context.Context, actor model.Profile, id string) (model.Batch, error) {
	if err := ctx.Err(); err != nil {
		return model.Batch{}, err
	}
	if !permission.Allows(actor, permission.View) {
		return model.Batch{}, model.ErrUnauthorized
	}
	return svc.Store.GetBatch(id)
}

func (svc *BatchService) UpdateBatch(ctx context.Context, actor model.Profile, candidate model.Batch) (model.Batch, error) {
	if err := ctx.Err(); err != nil {
		return model.Batch{}, err
	}
	if !permission.Allows(actor, permission.Edit) {
		return model.Batch{}, model.ErrUnauthorized
	}
	current, err := svc.Store.GetBatch(candidate.ID)
	if err != nil {
		return model.Batch{}, err
	}
	if candidate.Version != current.Version {
		return model.Batch{}, model.ErrConflict
	}
	if candidate.TotalStock < current.Issued {
		return model.Batch{}, errors.New("stock cannot be below issued quantity")
	}
	candidate.Remaining = candidate.TotalStock - current.Issued
	candidate.Issued = current.Issued
	candidate.CreatedAt, candidate.CreatedBy = current.CreatedAt, current.CreatedBy
	candidate.UpdatedAt, candidate.UpdatedBy = svc.now(), actor.ID
	candidate.Version++
	candidate.Tags = model.NormalizeTags(candidate.Tags)
	if err := candidate.Validate(svc.now()); err != nil {
		return model.Batch{}, err
	}
	if err := svc.Store.SaveBatch(candidate); err != nil {
		return model.Batch{}, err
	}
	_ = svc.recordAudit(candidate.ID, "edit", actor.ID, "batch updated")
	return candidate, nil
}

func (svc *BatchService) DeleteBatch(ctx context.Context, actor model.Profile, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !permission.Allows(actor, permission.Delete) {
		return model.ErrUnauthorized
	}
	batch, err := svc.Store.GetBatch(id)
	if err != nil {
		return err
	}
	if batch.Issued > 0 {
		return errors.New("issued batch cannot be deleted")
	}
	if err := svc.Store.DeleteBatch(id); err != nil {
		return err
	}
	_ = svc.recordAudit(id, "delete", actor.ID, "batch deleted")
	return nil
}

func (svc *BatchService) ListBatches(ctx context.Context, actor model.Profile, filter model.BatchFilter, limit int) (model.BatchPage, error) {
	if err := ctx.Err(); err != nil {
		return model.BatchPage{}, err
	}
	if !permission.Allows(actor, permission.View) {
		return model.BatchPage{}, model.ErrUnauthorized
	}
	items, err := svc.Store.ListBatches()
	if err != nil {
		return model.BatchPage{}, err
	}
	filtered := make([]model.Batch, 0, len(items))
	for _, item := range items {
		if filter.Match(item, svc.now()) {
			filtered = append(filtered, item)
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].UpdatedAt.After(filtered[j].UpdatedAt) })
	if limit <= 0 || limit > len(filtered) {
		limit = len(filtered)
	}
	page := model.BatchPage{Items: filtered[:limit], Total: len(filtered)}
	if limit < len(filtered) {
		page.NextCursor = filtered[limit-1].ID
	}
	return page, nil
}

func (svc *BatchService) ExportBatches(ctx context.Context, actor model.Profile, filter model.BatchFilter) (string, error) {
	if !permission.Allows(actor, permission.Export) {
		return "", model.ErrUnauthorized
	}
	page, err := svc.ListBatches(ctx, actor, filter, 0)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	builder.WriteString("id,name,amount_cents,currency,total_stock,remaining,issued,start_at,end_at,status\n")
	for _, item := range page.Items {
		fmt.Fprintf(&builder, "%s,%q,%d,%s,%d,%d,%d,%s,%s,%s\n", item.ID, item.Name, item.AmountCents, item.Currency, item.TotalStock, item.Remaining, item.Issued, item.StartAt.Format(time.RFC3339), item.EndAt.Format(time.RFC3339), item.Status)
	}
	_ = svc.recordAudit("export", "export", actor.ID, fmt.Sprintf("exported %d batches", len(page.Items)))
	return builder.String(), nil
}

func (svc *BatchService) ActivateBatch(ctx context.Context, actor model.Profile, id string) (model.Batch, error) {
	batch, err := svc.Store.GetBatch(id)
	if err != nil {
		return model.Batch{}, err
	}
	batch.Status = model.StatusActive
	return svc.UpdateBatch(ctx, actor, batch)
}

func (svc *BatchService) PauseBatch(ctx context.Context, actor model.Profile, id string) (model.Batch, error) {
	batch, err := svc.Store.GetBatch(id)
	if err != nil {
		return model.Batch{}, err
	}
	batch.Status = model.StatusPaused
	return svc.UpdateBatch(ctx, actor, batch)
}

func (svc *BatchService) recordAudit(entityID, action, actor, details string) error {
	return svc.Store.SaveAudit(model.Audit{ID: newID("aud"), Entity: "Batch", EntityID: entityID, Action: action, Actor: actor, Details: details, CreatedAt: svc.now()})
}

func newID(prefix string) string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(buf)
}

func withBatchTx(s *store.Store, fn func(*bbolt.Tx) error) error {
	return s.Update(fn)
}
