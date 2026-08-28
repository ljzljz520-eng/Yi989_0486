package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"couponbatch/internal/model"
)

type Idempotency struct {
	mu      sync.Mutex
	entries map[string]idempotentEntry
	ttl     time.Duration
}

type idempotentEntry struct {
	Record  model.Record
	Expires time.Time
}

func NewIdempotency(ttl time.Duration) *Idempotency {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	return &Idempotency{entries: make(map[string]idempotentEntry), ttl: ttl}
}

func (cache *Idempotency) Get(key string, now time.Time) (model.Record, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	entry, ok := cache.entries[key]
	if !ok {
		return model.Record{}, false
	}
	if now.After(entry.Expires) {
		delete(cache.entries, key)
		return model.Record{}, false
	}
	return entry.Record, true
}

func (cache *Idempotency) Put(key string, record model.Record, now time.Time) error {
	if key == "" {
		return errors.New("idempotency key is required")
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.entries[key] = idempotentEntry{Record: record, Expires: now.Add(cache.ttl)}
	return nil
}

func (svc *ClaimService) ClaimIdempotent(ctx context.Context, actor model.Profile, batchID, customerID, key string, cache *Idempotency) (model.Record, error) {
	now := svc.now()
	if record, ok := cache.Get(key, now); ok {
		return record, nil
	}
	record, err := svc.Claim(ctx, actor, batchID, customerID)
	if err != nil {
		return model.Record{}, err
	}
	if err := cache.Put(key, record, now); err != nil {
		return model.Record{}, err
	}
	return record, nil
}

func (cache *Idempotency) Purge(now time.Time) int {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	removed := 0
	for key, entry := range cache.entries {
		if now.After(entry.Expires) {
			delete(cache.entries, key)
			removed++
		}
	}
	return removed
}
