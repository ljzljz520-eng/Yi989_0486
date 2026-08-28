package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"couponbatch/internal/model"
	"couponbatch/internal/permission"
	"couponbatch/internal/store"
)

func fixture(t *testing.T) (*store.Store, model.Profile, *BatchService, *ClaimService) {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "service.db"))
	if err != nil {
		t.Fatal(err)
	}
	actor := permission.DefaultAdmin("admin")
	batches, claims := NewBatchService(s), NewClaimService(s)
	return s, actor, batches, claims
}

func TestBatchServiceCRUD(t *testing.T) {
	s, actor, batches, _ := fixture(t)
	defer s.Close()
	now := time.Now()
	created, err := batches.CreateBatch(context.Background(), actor, model.Batch{Name: "CRUD", AmountCents: 250, Currency: "CNY", TotalStock: 5, StartAt: now.Add(-time.Hour), EndAt: now.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	created.Description = "updated"
	updated, err := batches.UpdateBatch(context.Background(), actor, created)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Version != 2 {
		t.Fatal(updated.Version)
	}
	page, err := batches.ListBatches(context.Background(), actor, model.BatchFilter{Query: "updated"}, 0)
	if err != nil || len(page.Items) != 1 {
		t.Fatalf("%v %v", page, err)
	}
	csv, err := batches.ExportBatches(context.Background(), actor, model.BatchFilter{})
	if err != nil || len(csv) < 10 {
		t.Fatalf("%q %v", csv, err)
	}
	if err := batches.DeleteBatch(context.Background(), actor, created.ID); err != nil {
		t.Fatal(err)
	}
}

func TestClaimLifecycle(t *testing.T) {
	s, actor, batches, claims := fixture(t)
	defer s.Close()
	now := time.Now()
	batch, err := batches.CreateBatch(context.Background(), actor, model.Batch{ID: "claim-batch", Name: "Claim", AmountCents: 300, Currency: "CNY", TotalStock: 2, StartAt: now.Add(-time.Hour), EndAt: now.Add(time.Hour), Status: model.StatusActive})
	if err != nil {
		t.Fatal(err)
	}
	record, err := claims.Claim(context.Background(), actor, batch.ID, "customer")
	if err != nil {
		t.Fatal(err)
	}
	confirmed, err := claims.Confirm(context.Background(), actor, record.ID)
	if err != nil || confirmed.State != model.RecordConfirmed {
		t.Fatalf("%v %v", confirmed, err)
	}
	reclaimed, err := claims.Reclaim(context.Background(), actor, record.ID, "return")
	if err != nil || reclaimed.State != model.RecordReclaimed {
		t.Fatalf("%v %v", reclaimed, err)
	}
	if got, _ := batches.GetBatch(context.Background(), actor, batch.ID); got.Remaining != 2 {
		t.Fatal(got.Remaining)
	}
}

func TestServiceRejectsUnauthorized(t *testing.T) {
	s, _, batches, _ := fixture(t)
	defer s.Close()
	_, err := batches.ListBatches(context.Background(), model.Profile{ID: "off", Active: false}, model.BatchFilter{}, 0)
	if err != model.ErrUnauthorized {
		t.Fatal(err)
	}
}
