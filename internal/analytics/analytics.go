package analytics

import (
	"sort"
	"strings"
	"time"

	"couponbatch/internal/model"
)

type TimelinePoint struct {
	Date      string `json:"date"`
	Created   int    `json:"created"`
	Confirmed int    `json:"confirmed"`
	Reclaimed int    `json:"reclaimed"`
}

type CustomerStat struct {
	CustomerID  string `json:"customer_id"`
	Claims      int    `json:"claims"`
	Confirmed   int    `json:"confirmed"`
	Reclaimed   int    `json:"reclaimed"`
	AmountCents int64  `json:"amount_cents"`
}

type StatusCount struct {
	Status model.RecordState `json:"status"`
	Count  int               `json:"count"`
}

func BuildTimeline(records []model.Record, from, to time.Time) []TimelinePoint {
	if to.Before(from) {
		return nil
	}
	points := make(map[string]*TimelinePoint)
	for day := from.UTC().Truncate(24 * time.Hour); !day.After(to); day = day.Add(24 * time.Hour) {
		key := day.Format("2006-01-02")
		points[key] = &TimelinePoint{Date: key}
	}
	for _, record := range records {
		key := record.CreatedAt.UTC().Format("2006-01-02")
		point, ok := points[key]
		if !ok {
			continue
		}
		point.Created++
		switch record.State {
		case model.RecordConfirmed:
			point.Confirmed++
		case model.RecordReclaimed:
			point.Reclaimed++
		}
	}
	result := make([]TimelinePoint, 0, len(points))
	for _, point := range points {
		result = append(result, *point)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Date < result[j].Date })
	return result
}

func CustomerStats(records []model.Record) []CustomerStat {
	stats := make(map[string]*CustomerStat)
	for _, record := range records {
		key := strings.TrimSpace(record.CustomerID)
		if key == "" {
			continue
		}
		stat, ok := stats[key]
		if !ok {
			stat = &CustomerStat{CustomerID: key}
			stats[key] = stat
		}
		stat.Claims++
		stat.AmountCents += record.AmountCents
		if record.State == model.RecordConfirmed {
			stat.Confirmed++
		}
		if record.State == model.RecordReclaimed {
			stat.Reclaimed++
		}
	}
	result := make([]CustomerStat, 0, len(stats))
	for _, stat := range stats {
		result = append(result, *stat)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Claims == result[j].Claims {
			return result[i].CustomerID < result[j].CustomerID
		}
		return result[i].Claims > result[j].Claims
	})
	return result
}

func StateCounts(records []model.Record) []StatusCount {
	counts := make(map[model.RecordState]int)
	for _, record := range records {
		counts[record.State]++
	}
	result := make([]StatusCount, 0, len(counts))
	for state, count := range counts {
		result = append(result, StatusCount{Status: state, Count: count})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Status < result[j].Status })
	return result
}

func AvailabilityRate(batches []model.Batch) float64 {
	if len(batches) == 0 {
		return 0
	}
	available := 0
	for _, batch := range batches {
		if batch.Remaining > 0 && batch.TotalStock > 0 {
			available++
		}
	}
	return float64(available) / float64(len(batches))
}

func IssuanceRate(batch model.Batch) float64 {
	if batch.TotalStock <= 0 {
		return 0
	}
	return float64(batch.Issued) / float64(batch.TotalStock)
}

func GroupByStatus(batches []model.Batch) map[model.BatchStatus][]model.Batch {
	groups := make(map[model.BatchStatus][]model.Batch)
	for _, batch := range batches {
		groups[batch.Status] = append(groups[batch.Status], batch)
	}
	for status := range groups {
		sort.Slice(groups[status], func(i, j int) bool { return groups[status][i].Name < groups[status][j].Name })
	}
	return groups
}

func TopBatches(batches []model.Batch, limit int) []model.Batch {
	copyItems := append([]model.Batch(nil), batches...)
	sort.Slice(copyItems, func(i, j int) bool { return copyItems[i].Issued > copyItems[j].Issued })
	if limit <= 0 || limit >= len(copyItems) {
		return copyItems
	}
	return copyItems[:limit]
}
