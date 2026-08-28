package analytics

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"couponbatch/internal/model"
)

type Dashboard struct {
	Title    string          `json:"title"`
	Subtitle string          `json:"subtitle"`
	Cards    []Card          `json:"cards"`
	Status   []StatusCount   `json:"status"`
	Timeline []TimelinePoint `json:"timeline"`
	Top      []model.Batch   `json:"top"`
}

type Card struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Hint  string `json:"hint"`
}

func DashboardFor(batches []model.Batch, records []model.Record) Dashboard {
	groups := GroupByStatus(batches)
	confirmed, reclaimed := 0, 0
	for _, record := range records {
		if record.State == model.RecordConfirmed {
			confirmed++
		}
		if record.State == model.RecordReclaimed {
			reclaimed++
		}
	}
	return Dashboard{Title: "Coupon batches", Subtitle: "Inventory and claim operations", Cards: []Card{
		{Label: "Batches", Value: fmt.Sprintf("%d", len(batches)), Hint: fmt.Sprintf("%d active", len(groups[model.StatusActive]))},
		{Label: "Remaining", Value: fmt.Sprintf("%d", remainingStock(batches)), Hint: fmt.Sprintf("%.0f%% available", AvailabilityRate(batches)*100)},
		{Label: "Confirmed", Value: fmt.Sprintf("%d", confirmed), Hint: fmt.Sprintf("%d reclaimed", reclaimed)},
	}, Status: StateCounts(records), Top: TopBatches(batches, 5)}
}

func remainingStock(batches []model.Batch) int {
	total := 0
	for _, batch := range batches {
		total += batch.Remaining
	}
	return total
}

func RenderText(dashboard Dashboard) string {
	var builder strings.Builder
	builder.WriteString(dashboard.Title + "\n")
	builder.WriteString(dashboard.Subtitle + "\n")
	for _, card := range dashboard.Cards {
		builder.WriteString(card.Label + ": " + card.Value + " (" + card.Hint + ")\n")
	}
	return builder.String()
}

func RenderJSON(dashboard Dashboard) ([]byte, error) { return json.MarshalIndent(dashboard, "", "  ") }

func RankCustomers(records []model.Record) []CustomerStat {
	items := CustomerStats(records)
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].AmountCents == items[j].AmountCents {
			return items[i].CustomerID < items[j].CustomerID
		}
		return items[i].AmountCents > items[j].AmountCents
	})
	return items
}

func ComparePeriods(current, previous []model.Record) map[string]int {
	currentConfirmed, previousConfirmed := 0, 0
	for _, record := range current {
		if record.State == model.RecordConfirmed {
			currentConfirmed++
		}
	}
	for _, record := range previous {
		if record.State == model.RecordConfirmed {
			previousConfirmed++
		}
	}
	return map[string]int{"current_claims": len(current), "previous_claims": len(previous), "current_confirmed": currentConfirmed, "previous_confirmed": previousConfirmed, "confirmed_delta": currentConfirmed - previousConfirmed}
}

func FilterRecords(records []model.Record, state model.RecordState, customer string) []model.Record {
	result := make([]model.Record, 0, len(records))
	for _, record := range records {
		if state != "" && record.State != state {
			continue
		}
		if customer != "" && !strings.EqualFold(record.CustomerID, customer) {
			continue
		}
		result = append(result, record)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.Before(result[j].CreatedAt) })
	return result
}

func MergeTimeline(points ...[]TimelinePoint) []TimelinePoint {
	merged := make(map[string]TimelinePoint)
	for _, series := range points {
		for _, point := range series {
			value := merged[point.Date]
			value.Date = point.Date
			value.Created += point.Created
			value.Confirmed += point.Confirmed
			value.Reclaimed += point.Reclaimed
			merged[point.Date] = value
		}
	}
	result := make([]TimelinePoint, 0, len(merged))
	for _, point := range merged {
		result = append(result, point)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Date < result[j].Date })
	return result
}
