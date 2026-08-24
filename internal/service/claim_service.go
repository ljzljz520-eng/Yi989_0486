package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"couponbatch/internal/model"
	"couponbatch/internal/permission"
	"couponbatch/internal/store"
	"go.etcd.io/bbolt"
)

type ClaimService struct {
	Store *store.Store
	Now   Clock
}

func NewClaimService(s *store.Store) *ClaimService { return &ClaimService{Store: s, Now: time.Now} }

func (svc *ClaimService) now() time.Time {
	if svc.Now == nil {
		return time.Now()
	}
	return svc.Now()
}

func (svc *ClaimService) Claim(ctx context.Context, actor model.Profile, batchID, customerID string) (model.Record, error) {
	if err := ctx.Err(); err != nil {
		return model.Record{}, err
	}
	if !permission.Allows(actor, permission.Claim) {
		return model.Record{}, model.ErrUnauthorized
	}
	if customerID == "" {
		return model.Record{}, model.ErrInvalidRequest
	}
	var result model.Record
	err := svc.Store.Update(func(tx *bbolt.Tx) error {
		batch, err := getBatchTx(tx, batchID)
		if err != nil {
			return err
		}
		if !batch.IsAvailable(svc.now()) {
			return errors.New("batch is not available")
		}
		batch.Remaining--
		batch.Issued++
		batch.Version++
		batch.UpdatedAt = svc.now()
		if err := updateBatchTx(tx, batch); err != nil {
			return err
		}
		result = model.Record{ID: newID("rec"), BatchID: batchID, CustomerID: customerID, State: model.RecordReserved, AmountCents: batch.AmountCents, CreatedAt: svc.now(), UpdatedAt: svc.now(), Version: 1}
		return updateRecordTx(tx, result)
	})
	if err != nil {
		return model.Record{}, err
	}
	return result, nil
}

func (svc *ClaimService) Confirm(ctx context.Context, actor model.Profile, recordID string) (model.Record, error) {
	if err := ctx.Err(); err != nil {
		return model.Record{}, err
	}
	if !permission.Allows(actor, permission.Claim) {
		return model.Record{}, model.ErrUnauthorized
	}
	var result model.Record
	err := svc.Store.Update(func(tx *bbolt.Tx) error {
		record, err := getRecordTx(tx, recordID)
		if err != nil {
			return err
		}
		if record.State != model.RecordReserved {
			return model.ErrInvalidState
		}
		now := svc.now()
		record.State = model.RecordConfirmed
		record.ConfirmedAt = &now
		record.UpdatedAt = now
		record.Version++
		if err := updateRecordTx(tx, record); err != nil {
			return err
		}
		result = record
		return nil
	})
	return result, err
}

func (svc *ClaimService) Reclaim(ctx context.Context, actor model.Profile, recordID, reason string) (model.Record, error) {
	if err := ctx.Err(); err != nil {
		return model.Record{}, err
	}
	if !permission.Allows(actor, permission.Claim) {
		return model.Record{}, model.ErrUnauthorized
	}
	var result model.Record
	err := svc.Store.Update(func(tx *bbolt.Tx) error {
		record, err := getRecordTx(tx, recordID)
		if err != nil {
			return err
		}
		if record.State != model.RecordConfirmed && record.State != model.RecordReserved {
			return model.ErrInvalidState
		}
		batch, err := getBatchTx(tx, record.BatchID)
		if err != nil {
			return err
		}
		batch.Remaining++
		if batch.Issued > 0 {
			batch.Issued--
		}
		batch.Version++
		batch.UpdatedAt = svc.now()
		if err := updateBatchTx(tx, batch); err != nil {
			return err
		}
		now := svc.now()
		record.State = model.RecordReclaimed
		record.ReclaimedAt = &now
		record.UpdatedAt = now
		record.Reason = reason
		record.Version++
		if err := updateRecordTx(tx, record); err != nil {
			return err
		}
		result = record
		return nil
	})
	return result, err
}

func (svc *ClaimService) Cancel(ctx context.Context, actor model.Profile, recordID string) (model.Record, error) {
	if err := ctx.Err(); err != nil {
		return model.Record{}, err
	}
	if !permission.Allows(actor, permission.Claim) {
		return model.Record{}, model.ErrUnauthorized
	}
	record, err := svc.Store.GetRecord(recordID)
	if err != nil {
		return model.Record{}, err
	}
	if record.State != model.RecordReserved {
		return model.Record{}, model.ErrInvalidState
	}
	return svc.Reclaim(ctx, actor, recordID, "cancelled")
}

func (svc *ClaimService) GetRecord(ctx context.Context, actor model.Profile, recordID string) (model.Record, error) {
	if err := ctx.Err(); err != nil {
		return model.Record{}, err
	}
	if !permission.Allows(actor, permission.View) {
		return model.Record{}, model.ErrUnauthorized
	}
	return svc.Store.GetRecord(recordID)
}

func (svc *ClaimService) AvailableBatches(ctx context.Context, actor model.Profile) ([]model.Batch, error) {
	page, err := NewBatchService(svc.Store).ListBatches(ctx, actor, model.BatchFilter{Available: true}, 0)
	return page.Items, err
}

func getBatchTx(tx *bbolt.Tx, id string) (model.Batch, error) {
	var batch model.Batch
	b := tx.Bucket([]byte("batches"))
	if b == nil || b.Get([]byte(id)) == nil {
		return model.Batch{}, model.ErrNotFound
	}
	if err := model.Decode(b.Get([]byte(id)), &batch); err != nil {
		return model.Batch{}, err
	}
	return batch, nil
}

func updateBatchTx(tx *bbolt.Tx, batch model.Batch) error {
	payload, err := model.Encode(batch)
	if err != nil {
		return err
	}
	return tx.Bucket([]byte("batches")).Put([]byte(batch.ID), payload)
}

func getRecordTx(tx *bbolt.Tx, id string) (model.Record, error) {
	var record model.Record
	b := tx.Bucket([]byte("records"))
	if b == nil || b.Get([]byte(id)) == nil {
		return model.Record{}, model.ErrNotFound
	}
	if err := model.Decode(b.Get([]byte(id)), &record); err != nil {
		return model.Record{}, err
	}
	return record, nil
}

func updateRecordTx(tx *bbolt.Tx, record model.Record) error {
	payload, err := model.Encode(record)
	if err != nil {
		return fmt.Errorf("encode record: %w", err)
	}
	return tx.Bucket([]byte("records")).Put([]byte(record.ID), payload)
}
