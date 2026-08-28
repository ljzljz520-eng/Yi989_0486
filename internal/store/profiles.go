package store

import (
	"couponbatch/internal/model"
	"go.etcd.io/bbolt"
)

func (s *Store) SaveProfile(profile model.Profile) error {
	return s.Update(func(tx *bbolt.Tx) error { return putJSON(tx, profileBucket, profile.ID, profile) })
}

func (s *Store) GetProfile(id string) (model.Profile, error) {
	var profile model.Profile
	err := s.View(func(tx *bbolt.Tx) error { return getJSON(tx, profileBucket, id, &profile) })
	if err != nil {
		return model.Profile{}, model.ErrNotFound
	}
	return profile, nil
}

func (s *Store) ListProfiles() ([]model.Profile, error) {
	items := make([]model.Profile, 0)
	err := s.View(func(tx *bbolt.Tx) error {
		return scanJSON(tx, profileBucket, func(_ string, item model.Profile) bool {
			items = append(items, item)
			return true
		})
	})
	return items, err
}
