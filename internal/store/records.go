package store

import (
	"couponbatch/internal/model"
	"go.etcd.io/bbolt"
)

func (s *Store) SaveRecord(record model.Record) error {
	return s.Update(func(tx *bbolt.Tx) error { return putJSON(tx, recordBucket, record.ID, record) })
}

func (s *Store) GetRecord(id string) (model.Record, error) {
	var record model.Record
	err := s.View(func(tx *bbolt.Tx) error { return getJSON(tx, recordBucket, id, &record) })
	if err != nil {
		return model.Record{}, model.ErrNotFound
	}
	return record, nil
}

func (s *Store) DeleteRecord(id string) error {
	return s.Update(func(tx *bbolt.Tx) error { return deleteKey(tx, recordBucket, id) })
}

func (s *Store) UpdateRecordTx(tx *bbolt.Tx, record model.Record) error {
	return putJSON(tx, recordBucket, record.ID, record)
}

func (s *Store) ListRecords() ([]model.Record, error) {
	items := make([]model.Record, 0)
	err := s.View(func(tx *bbolt.Tx) error {
		return scanJSON(tx, recordBucket, func(_ string, item model.Record) bool {
			items = append(items, item)
			return true
		})
	})
	return items, err
}
