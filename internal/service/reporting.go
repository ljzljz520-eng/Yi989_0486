package service

import (
	"context"
	"sort"
	"strings"
	"time"

	"couponbatch/internal/model"
	"couponbatch/internal/permission"
)

type InventorySummary struct {
	TotalBatches  int   `json:"total_batches"`
	ActiveBatches int   `json:"active_batches"`
	TotalStock    int   `json:"total_stock"`
	Remaining     int   `json:"remaining"`
	Issued        int   `json:"issued"`
	AmountCents   int64 `json:"amount_cents"`
}

func (svc *BatchService) Summary(ctx context.Context, actor model.Profile, filter model.BatchFilter) (InventorySummary, error) {
	page, err := svc.ListBatches(ctx, actor, filter, 0)
	if err != nil {
		return InventorySummary{}, err
	}
	var summary InventorySummary
	for _, item := range page.Items {
		summary.TotalBatches++
		summary.TotalStock += item.TotalStock
		summary.Remaining += item.Remaining
		summary.Issued += item.Issued
		summary.AmountCents += item.AmountCents
		if item.Status == model.StatusActive {
			summary.ActiveBatches++
		}
	}
	return summary, nil
}

func (svc *BatchService) Search(ctx context.Context, actor model.Profile, query string) ([]model.Batch, error) {
	page, err := svc.ListBatches(ctx, actor, model.BatchFilter{Query: query}, 0)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (svc *BatchService) Upcoming(ctx context.Context, actor model.Profile, within time.Duration) ([]model.Batch, error) {
	now := svc.now()
	page, err := svc.ListBatches(ctx, actor, model.BatchFilter{StartsAfter: now, EndsBefore: now.Add(within)}, 0)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (svc *BatchService) Tags(ctx context.Context, actor model.Profile) ([]string, error) {
	page, err := svc.ListBatches(ctx, actor, model.BatchFilter{}, 0)
	if err != nil {
		return nil, err
	}
	set := make(map[string]struct{})
	for _, item := range page.Items {
		for _, tag := range item.Tags {
			set[tag] = struct{}{}
		}
	}
	result := make([]string, 0, len(set))
	for tag := range set {
		result = append(result, tag)
	}
	sort.Strings(result)
	return result, nil
}

func (svc *BatchService) CanExport(actor model.Profile) bool {
	return permission.Allows(actor, permission.Export)
}

func SearchTerms(query string) []string {
	parts := strings.Fields(strings.ToLower(query))
	result := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		if _, ok := seen[part]; !ok {
			seen[part] = struct{}{}
			result = append(result, part)
		}
	}
	return result
}
