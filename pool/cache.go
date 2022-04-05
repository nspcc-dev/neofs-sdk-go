package pool

import (
	"strings"
	"sync/atomic"

	lru "github.com/hashicorp/golang-lru"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

type sessionCache struct {
	cache        *lru.Cache
	currentEpoch uint64
}

type cacheValue struct {
	token *session.Token
}

func newCache() (*sessionCache, error) {
	cache, err := lru.New(100)
	if err != nil {
		return nil, err
	}

	return &sessionCache{cache: cache}, nil
}

// Get returns a copy of the session token from the cache without signature
// and context related fields. Returns nil if token is missing in the cache.
// It is safe to modify and re-sign returned session token.
func (c *sessionCache) Get(key string) *session.Token {
	valueRaw, ok := c.cache.Get(key)
	if !ok {
		return nil
	}

	value := valueRaw.(*cacheValue)
	if c.expired(value) {
		c.cache.Remove(key)
		return nil
	}

	if value.token == nil {
		return nil
	}

	res := copySessionTokenWithoutSignatureAndContext(*value.token)

	return &res
}

func (c *sessionCache) Put(key string, token *session.Token) bool {
	return c.cache.Add(key, &cacheValue{
		token: token,
	})
}

func (c *sessionCache) DeleteByPrefix(prefix string) {
	for _, key := range c.cache.Keys() {
		if strings.HasPrefix(key.(string), prefix) {
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

func (c *sessionCache) expired(val *cacheValue) bool {
	epoch := atomic.LoadUint64(&c.currentEpoch)
	return val.token.Exp() <= epoch
}
