package tracers

import (
	"sync"
	"time"
)

const (
	cacheCheckInterval = time.Minute
	cacheItemsTTL      = time.Minute * 30
)

// TxTraceResult is the result of a single transaction trace.
type TxTraceResult struct {
	Result interface{} `json:"result,omitempty"` // Trace results produced by the tracer
	Error  string      `json:"error,omitempty"`  // Trace failure produced by the tracer
}

type cachedTraceResults struct {
	Trace     []*TxTraceResult
	CreatedAt time.Time
}

type TxTraceResultsCache struct {
	rwm sync.RWMutex

	// items is a map of block hash hex strings to the cached trace results for those blocks.
	items map[string]cachedTraceResults

	// ttl is the time-to-live for the items in the cache.
	ttl time.Duration
}

// TraceResultsCache stores slices of TxTraceResults on a per-block basis. Cached trace
// results are automatically cleared after 30 minutes.
var TraceResultsCache = &TxTraceResultsCache{
	items: make(map[string]cachedTraceResults),
}

func init() {
	go func() {
		for {
			time.Sleep(cacheCheckInterval)
			TraceResultsCache.deleteExpired()
		}
	}()
}

// Get returns a slice of TxTraceResults keyed by the hash of the block used to
// generate them. Returns nil if not the trace results for the given block are not found.
func (cache *TxTraceResultsCache) Get(blockHashHex string) []*TxTraceResult {
	cache.rwm.RLock()
	defer cache.rwm.RUnlock()

	cachedTrace := cache.items[blockHashHex].Trace
	delete(cache.items, blockHashHex)
	return cachedTrace
}

// Set caches a slice of TxTraceResults for a given block hash.
func (cache *TxTraceResultsCache) Set(blockHashHex string, value []*TxTraceResult) {
	cache.rwm.Lock()
	defer cache.rwm.Unlock()

	cache.items[blockHashHex] = cachedTraceResults{
		Trace:     value,
		CreatedAt: time.Now(),
	}
}

func (cache *TxTraceResultsCache) deleteExpired() {
	cache.rwm.Lock()
	defer cache.rwm.Unlock()

	now := time.Now()
	for k, v := range cache.items {
		if now.After(v.CreatedAt.Add(cacheItemsTTL)) {
			delete(cache.items, k)
		}
	}
}
