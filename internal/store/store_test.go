package store

import (
	"path/filepath"
	"testing"
	"time"

	"couponbatch/internal/model"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "coupon.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	batch := model.Batch{ID: "persist-batch", Name: "Persist", AmountCents: 500, Currency: "CNY", TotalStock: 4, Remaining: 4, StartAt: now.Add(-time.Hour), EndAt: now.Add(time.Hour), Status: model.StatusActive}
	if err := s.SaveBatch(batch); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveProfile(model.Profile{ID: "p", Active: true}); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveAudit(model.Audit{ID: "a", EntityID: batch.ID}); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveRecord(model.Record{ID: "r", BatchID: batch.ID, CustomerID: "c", State: model.RecordReserved, AmountCents: 500}); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if got, err := s.GetBatch(batch.ID); err != nil || got.Name != batch.Name {
		t.Fatalf("batch %#v %v", got, err)
	}
	if _, err := s.GetProfile("p"); err != nil {
		t.Fatal(err)
	}
	if records, err := s.ListRecords(); err != nil || len(records) != 1 {
		t.Fatalf("records %v %v", records, err)
	}
}

func TestStoreBuckets(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "x.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.ListBatches(); err != nil {
		t.Fatal(err)
	}
}
