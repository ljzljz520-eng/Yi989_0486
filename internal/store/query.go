package store

import (
	"strings"
	"time"

	"couponbatch/internal/model"
	"go.etcd.io/bbolt"
)

func (s *Store) FindBatches(filter model.BatchFilter, now time.Time) ([]model.Batch, error) {
	items := make([]model.Batch, 0)
	err := s.View(func(tx *bbolt.Tx) error {
		return scanJSON(tx, batchBucket, func(_ string, item model.Batch) bool {
			if filter.Match(item, now) {
				items = append(items, item)
			}
			return true
		})
	})
	model.SortByCursor(items)
	return items, err
}

func (s *Store) FindRecords(batchID, customerID string) ([]model.Record, error) {
	items := make([]model.Record, 0)
	err := s.View(func(tx *bbolt.Tx) error {
		return scanJSON(tx, recordBucket, func(_ string, item model.Record) bool {
			if batchID != "" && item.BatchID != batchID {
				return true
			}
			if customerID != "" && !strings.EqualFold(item.CustomerID, customerID) {
				return true
			}
			items = append(items, item)
			return true
		})
	})
	return items, err
}

func (s *Store) Count(bucket []byte) (int, error) {
	count := 0
	err := s.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}
		count = b.Stats().KeyN
		return nil
	})
	return count, err
}

func (s *Store) Snapshot() (map[string]int, error) {
	result := make(map[string]int)
	for name, bucket := range map[string][]byte{"batches": batchBucket, "records": recordBucket, "audits": auditBucket, "profiles": profileBucket} {
		count, err := s.Count(bucket)
		if err != nil {
			return nil, err
		}
		result[name] = count
	}
	return result, nil
}
