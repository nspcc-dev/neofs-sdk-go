package pool

import (
	"strings"
	"sync/atomic"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessionv2 "github.com/nspcc-dev/neofs-sdk-go/session/v2"
)

const (
	defaultSessionCacheSize = 700
)

type sessionCache struct {
	cache        *lru.Cache[string, *cacheValue]
	currentEpoch uint64
}

type cacheValue struct {
	token   session.Object
	tokenV2 sessionv2.Token
	version int
}

func newCache(cacheSize int) (*sessionCache, error) {
	cache, err := lru.New[string, *cacheValue](cacheSize)
	if err != nil {
		return nil, err
	}

	return &sessionCache{cache: cache}, nil
}

// Get returns a copy of the session token from the cache without signature
// and context related fields. Returns nil if token is missing in the cache.
// It is safe to modify and re-sign returned session token.
func (c *sessionCache) Get(key string) (session.Object, bool) {
	value, ok := c.cache.Get(key)
	if !ok || value.version != 1 {
		return session.Object{}, false
	}

	if value.token.ExpiredAt(atomic.LoadUint64(&c.currentEpoch) + 1) {
		c.cache.Remove(key)
		return session.Object{}, false
	}

	return value.token, true
}

// GetV2 returns a copy of the session token v2 from the cache without signature
// and context related fields. Returns nil if token is missing in the cache.
// It is safe to modify and re-sign returned session token.
func (c *sessionCache) GetV2(key string) (sessionv2.Token, bool) {
	value, ok := c.cache.Get(key)
	if !ok || value.version != 2 {
		return sessionv2.Token{}, false
	}

	if !value.tokenV2.ValidAt(time.Now()) {
		c.cache.Remove(key)
		return sessionv2.Token{}, false
	}

	return value.tokenV2, true
}

// Put adds or overwrites a v1 session token in the cache.
// If a token (v1 or v2) already exists under this key, it will be replaced.
func (c *sessionCache) Put(key string, token session.Object) bool {
	return c.cache.Add(key, &cacheValue{
		token:   token,
		version: 1,
	})
}

// PutV2 adds or overwrites a v2 session token in the cache.
// If a token (v1 or v2) already exists under this key, it will be replaced.
func (c *sessionCache) PutV2(key string, token sessionv2.Token) bool {
	return c.cache.Add(key, &cacheValue{
		tokenV2: token,
		version: 2,
	})
}

func (c *sessionCache) DeleteByPrefix(prefix string) {
	for _, key := range c.cache.Keys() {
		if strings.HasPrefix(key, prefix) {
			c.cache.Remove(key)
		}
	}
}

func (c *sessionCache) updateEpoch(newEpoch uint64) {
	epoch := atomic.LoadUint64(&c.currentEpoch)
	if newEpoch > epoch {
		atomic.StoreUint64(&c.currentEpoch, newEpoch)
	}
}

// Purge removes all session keys.
func (c *sessionCache) Purge() {
	c.cache.Purge()
}
