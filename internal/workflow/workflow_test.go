package workflow

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"couponbatch/internal/model"
	"couponbatch/internal/permission"
	"couponbatch/internal/store"
)

func TestWorkflowAccept(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "flow.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	engine := New(s)
	actor := permission.DefaultAdmin("operator")
	now := time.Now()
	batch, err := engine.Accept(context.Background(), actor, model.Batch{Name: "Workflow", AmountCents: 100, Currency: "CNY", TotalStock: 3, StartAt: now.Add(-time.Minute), EndAt: now.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if batch.Status != model.StatusActive {
		t.Fatal(batch.Status)
	}
}

func TestWorkflowPublish(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "flow.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	engine := New(s)
	actor := permission.DefaultAdmin("operator")
	now := time.Now()
	batch, err := engine.Accept(context.Background(), actor, model.Batch{Name: "Workflow", AmountCents: 100, Currency: "CNY", TotalStock: 3, StartAt: now.Add(-time.Minute), EndAt: now.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	record, err := engine.Publish(context.Background(), actor, batch.ID, "customer")
	if err != nil {
		t.Fatal(err)
	}
	if record.State != model.RecordConfirmed {
		t.Fatal(record.State)
	}
	reclaimed, err := engine.Recover(context.Background(), actor, record.ID)
	if err != nil || reclaimed.State != model.RecordReclaimed {
		t.Fatalf("%v %v", reclaimed, err)
	}
}

func TestWorkflowReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flow.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	engine := New(s)
	actor := permission.DefaultAdmin("operator")
	now := time.Now()
	if _, err := engine.Accept(context.Background(), actor, model.Batch{ID: "reopen", Name: "Reopen", AmountCents: 100, Currency: "CNY", TotalStock: 1, StartAt: now.Add(-time.Minute), EndAt: now.Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	reopened, err := engine.Reopen(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Store.Close()
	if _, err := reopened.Batches.GetBatch(context.Background(), actor, "reopen"); err != nil {
		t.Fatal(err)
	}
}

func TestWorkflow28(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "bug.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	engine := New(s)
	actor := permission.DefaultAdmin("operator")
	now := time.Now()
	batch, err := engine.Accept(context.Background(), actor, model.Batch{Name: "Bug", AmountCents: 100, Currency: "CNY", TotalStock: 1, StartAt: now.Add(-time.Minute), EndAt: now.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	record, err := engine.Publish(context.Background(), actor, batch.ID, "customer")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.ConfirmThenReclaim(context.Background(), actor, record.ID); err != nil {
		t.Fatal(err)
	}
	got, err := engine.Claims.GetRecord(context.Background(), actor, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != model.RecordReclaimed {
		t.Fatalf("expected reclaimed state, got %s", got.State)
	}
}
