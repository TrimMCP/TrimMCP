package capability

import (
	"context"
	"sync"
	"time"

	"github.com/slimmcp/slimmcp/pkg/types"
)

// Cache provides thread-safe caching of server capabilities with TTL
type Cache struct {
	mu           sync.RWMutex
	capabilities map[string]*cacheEntry
	detector     Detector
	defaultTTL   time.Duration
}

// cacheEntry represents a cached capability with expiration
type cacheEntry struct {
	capabilities *types.ServerCapabilities
	expiresAt    time.Time
}

// NewCache creates a new capability cache
func NewCache(detector Detector, defaultTTL time.Duration) *Cache {
	if defaultTTL == 0 {
		defaultTTL = 1 * time.Hour
	}

	return &Cache{
		capabilities: make(map[string]*cacheEntry),
		detector:     detector,
		defaultTTL:   defaultTTL,
	}
}

// GetCapabilities retrieves capabilities for a server, using cache if available
func (c *Cache) GetCapabilities(ctx context.Context, serverURL string) (*types.ServerCapabilities, error) {
	// Check cache first
	c.mu.RLock()
	entry, exists := c.capabilities[serverURL]
	c.mu.RUnlock()

	if exists && time.Now().Before(entry.expiresAt) {
		// Cache hit and not expired
		return entry.capabilities, nil
	}

	// Cache miss or expired - detect capabilities
	capabilities, err := c.detector.DetectCapabilities(ctx, serverURL)
	if err != nil {
		// If we have stale cache, return it on error
		if exists {
			return entry.capabilities, nil
		}
		return nil, err
	}

	// Update cache
	c.SetCapabilities(serverURL, capabilities)

	return capabilities, nil
}

// SetCapabilities stores capabilities in the cache
func (c *Cache) SetCapabilities(serverURL string, capabilities *types.ServerCapabilities) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ttl := c.defaultTTL
	if capabilities.CacheTTL > 0 {
		ttl = capabilities.CacheTTL
	}

	c.capabilities[serverURL] = &cacheEntry{
		capabilities: capabilities,
		expiresAt:    time.Now().Add(ttl),
	}
}

// Invalidate removes a specific server from the cache
func (c *Cache) Invalidate(serverURL string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.capabilities, serverURL)
}

// InvalidateAll clears all cached capabilities
func (c *Cache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.capabilities = make(map[string]*cacheEntry)
}

// GetCachedOnly retrieves capabilities from cache without detection
func (c *Cache) GetCachedOnly(serverURL string) (*types.ServerCapabilities, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.capabilities[serverURL]
	if !exists || time.Now().After(entry.expiresAt) {
		return nil, false
	}

	return entry.capabilities, true
}

// Size returns the number of cached entries
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.capabilities)
}

// CleanExpired removes expired entries from the cache
func (c *Cache) CleanExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	removed := 0

	for serverURL, entry := range c.capabilities {
		if now.After(entry.expiresAt) {
			delete(c.capabilities, serverURL)
			removed++
		}
	}

	return removed
}

// StartCleanupRoutine starts a background goroutine to periodically clean expired entries
func (c *Cache) StartCleanupRoutine(ctx context.Context, interval time.Duration) {
	if interval == 0 {
		interval = 15 * time.Minute
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.CleanExpired()
			}
		}
	}()
}
