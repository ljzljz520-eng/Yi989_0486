package export

import (
	"strings"
	"testing"

	"couponbatch/internal/model"
)

func TestCSVAndSummary(t *testing.T) {
	data, err := CSV([]model.Batch{{ID: "b", Name: "Name", Status: model.StatusActive, Remaining: 2, TotalStock: 3}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(data, "id,name") {
		t.Fatal(data)
	}
	if !strings.Contains(Summary([]model.Batch{{Status: model.StatusActive, Remaining: 2}}), "active=1") {
		t.Fatal("summary")
	}
}
