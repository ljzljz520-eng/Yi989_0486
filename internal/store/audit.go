package store

import (
	"couponbatch/internal/model"
	"go.etcd.io/bbolt"
)

func (s *Store) SaveAudit(audit model.Audit) error {
	return s.Update(func(tx *bbolt.Tx) error { return putJSON(tx, auditBucket, audit.ID, audit) })
}

func (s *Store) ListAudits(entityID string) ([]model.Audit, error) {
	items := make([]model.Audit, 0)
	err := s.View(func(tx *bbolt.Tx) error {
		return scanJSON(tx, auditBucket, func(_ string, item model.Audit) bool {
			if entityID == "" || item.EntityID == entityID {
				items = append(items, item)
			}
			return true
		})
	})
	return items, err
}
