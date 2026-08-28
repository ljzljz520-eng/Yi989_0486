package store

import (
	"fmt"
	"sort"

	"couponbatch/internal/model"
	"go.etcd.io/bbolt"
)

func (s *Store) SaveBatch(batch model.Batch) error {
	return s.Update(func(tx *bbolt.Tx) error {
		return putJSON(tx, batchBucket, batch.ID, batch)
	})
}

func (s *Store) GetBatch(id string) (model.Batch, error) {
	var batch model.Batch
	err := s.View(func(tx *bbolt.Tx) error { return getJSON(tx, batchBucket, id, &batch) })
	if err != nil {
		return model.Batch{}, model.ErrNotFound
	}
	return batch, nil
}

func (s *Store) DeleteBatch(id string) error {
	return s.Update(func(tx *bbolt.Tx) error {
		if _, err := s.getBatchTx(tx, id); err != nil {
			return model.ErrNotFound
		}
		return deleteKey(tx, batchBucket, id)
	})
}

func (s *Store) getBatchTx(tx *bbolt.Tx, id string) (model.Batch, error) {
	var batch model.Batch
	if err := getJSON(tx, batchBucket, id, &batch); err != nil {
		return model.Batch{}, model.ErrNotFound
	}
	return batch, nil
}

func (s *Store) ListBatches() ([]model.Batch, error) {
	items := make([]model.Batch, 0)
	err := s.View(func(tx *bbolt.Tx) error {
		err := scanJSON(tx, batchBucket, func(_ string, item model.Batch) bool {
			items = append(items, item)
			return true
		})
		if err == errStopScan {
			return nil
		}
		return err
	})
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })
	return items, err
}

func (s *Store) UpdateBatchTx(tx *bbolt.Tx, batch model.Batch) error {
	if batch.ID == "" {
		return fmt.Errorf("batch id missing")
	}
	return putJSON(tx, batchBucket, batch.ID, batch)
}
