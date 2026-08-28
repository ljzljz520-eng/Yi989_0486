package model

import "sort"

type Cursor struct {
	UpdatedAt int64  `json:"updated_at"`
	ID        string `json:"id"`
}

func EncodeCursor(batch Batch) string {
	return batch.UpdatedAt.UTC().Format("20060102150405.000000000") + ":" + batch.ID
}

func SortByCursor(items []Batch) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].UpdatedAt.Equal(items[j].UpdatedAt) {
			return items[i].ID < items[j].ID
		}
		return items[i].UpdatedAt.Before(items[j].UpdatedAt)
	})
}

func Page(items []Batch, cursor string, limit int) BatchPage {
	if limit <= 0 {
		limit = 50
	}
	start := 0
	if cursor != "" {
		for index, item := range items {
			if EncodeCursor(item) == cursor {
				start = index + 1
				break
			}
		}
	}
	if start > len(items) {
		start = len(items)
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}
	page := BatchPage{Items: append([]Batch(nil), items[start:end]...), Total: len(items)}
	if end < len(items) && end > start {
		page.NextCursor = EncodeCursor(items[end-1])
	}
	return page
}
