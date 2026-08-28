package model

import (
	"testing"
	"time"
)

func TestBatchValidationAndFilter(t *testing.T) {
	now := time.Now()
	batch := Batch{ID: "b1", Name: "Spring", AmountCents: 100, Currency: "CNY", TotalStock: 5, Remaining: 5, StartAt: now.Add(-time.Hour), EndAt: now.Add(time.Hour), Status: StatusActive, Tags: []string{"a"}}
	if err := batch.Validate(now); err != nil {
		t.Fatal(err)
	}
	if !batch.IsAvailable(now) {
		t.Fatal("batch should be available")
	}
	if !(BatchFilter{Query: "spring", Available: true}).Match(batch, now) {
		t.Fatal("filter mismatch")
	}
	if (BatchFilter{Status: StatusPaused}).Match(batch, now) {
		t.Fatal("paused filter matched active")
	}
	if got := NormalizeTags([]string{"b", "a", "a"}); len(got) != 2 {
		t.Fatal(got)
	}
}

func TestStatusLabelsAndSort(t *testing.T) {
	items := []Batch{{ID: "1", Name: "B", AmountCents: 200}, {ID: "2", Name: "A", AmountCents: 100}}
	SortBatches(items, "name", false)
	if items[0].Name != "A" {
		t.Fatal(items)
	}
	if StatusLabel(StatusArchived) != "Archived" {
		t.Fatal("label")
	}
}
