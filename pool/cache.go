package pool

import (
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/nspcc-dev/neofs-sdk-go/session/v2"
)

const (
	defaultSessionCacheSize = 700
)

type sessionCache struct {
	cache *lru.Cache[string, *cacheValue]
}

type cacheValue struct {
	token session.Token
}

func newCache(cacheSize int) (*sessionCache, error) {
	cache, err := lru.New[string, *cacheValue](cacheSize)
	if err != nil {
		return nil, err
	}

	return &sessionCache{cache: cache}, nil
}

// Get returns a copy of the session token V2 from the cache without signature
// and context related fields. Returns nil if token is missing in the cache.
// It is safe to modify and re-sign returned session token.
func (c *sessionCache) Get(key string) (session.Token, bool) {
	value, ok := c.cache.Get(key)
	if !ok {
		return session.Token{}, false
	}

	if c.expired(value) {
		c.cache.Remove(key)
		return session.Token{}, false
	}

	return value.token, true
}

func (c *sessionCache) Put(key string, token session.Token) bool {
	return c.cache.Add(key, &cacheValue{
		token: token,
	})
}

func (c *sessionCache) DeleteByPrefix(prefix string) {
	for _, key := range c.cache.Keys() {
		if strings.HasPrefix(key, prefix) {
			c.cache.Remove(key)
		}
	}
}

func (c *sessionCache) expired(val *cacheValue) bool {
	currentTime := uint64(time.Now().Unix())
	return !val.token.ValidAt(currentTime)
}

// Purge removes all session keys.
func (c *sessionCache) Purge() {
	c.cache.Purge()
}
