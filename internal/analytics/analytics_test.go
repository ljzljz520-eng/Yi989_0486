package analytics

import (
	"couponbatch/internal/model"
	"testing"
	"time"
)

func TestAnalytics(t *testing.T) {
	now := time.Now()
	records := []model.Record{{CustomerID: "a", State: model.RecordConfirmed, AmountCents: 100, CreatedAt: now}, {CustomerID: "a", State: model.RecordReclaimed, AmountCents: 100, CreatedAt: now}}
	if len(CustomerStats(records)) != 1 {
		t.Fatal("customer stats")
	}
	if len(BuildTimeline(records, now.Add(-time.Hour), now.Add(time.Hour))) == 0 {
		t.Fatal("timeline")
	}
	if len(StateCounts(records)) != 2 {
		t.Fatal("states")
	}
}
