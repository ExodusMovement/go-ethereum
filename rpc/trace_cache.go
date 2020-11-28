package rpc

import (
	"sync"
	"time"
)

// txTraceResult is the result of a single transaction trace.
type TxTraceResult struct {
	Result interface{} `json:"result,omitempty"` // Trace results produced by the tracer
	Error  string      `json:"error,omitempty"`  // Trace failure produced by the tracer
}

var (
	cache map[string]CachedItem
	rwm   sync.RWMutex
	ttl   = (time.Minute * 30).Nanoseconds()
)

type CachedItem struct {
	Trace     []*TxTraceResult
	CreatedAt int64
}

func createCache() {
	cache = make(map[string]CachedItem)

	done := make(chan bool)
	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				deleteExpired()
			}
		}
	}()
}

func GetFromCache(key string) []*TxTraceResult {
	rwm.RLock()
	if cache == nil {
		createCache()
	}
	defer rwm.RUnlock()

	cached := cache[key]
	if cached.Trace != nil {
		cpy := make([]*TxTraceResult, len(cached.Trace))
		for i, trace := range cached.Trace {
			cpy[i] = &TxTraceResult{trace.Result, trace.Error}
		}

		delete(cache, key)

		return cpy
	}

	return cache[key].Trace
}

func SetToCache(key string, value []*TxTraceResult) {
	rwm.Lock()
	if cache == nil {
		createCache()
	}
	defer rwm.Unlock()

	cache[key] = CachedItem{Trace: value, CreatedAt: time.Now().UnixNano()}
}

func deleteExpired() {
	now := time.Now().UnixNano()
	rwm.Lock()
	for k, v := range cache {
		if now > v.CreatedAt+ttl {
			delete(cache, k)
		}
	}
	defer rwm.Unlock()
}
