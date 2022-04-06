package pool

import (
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

type sessionCache struct {
	cache *lru.Cache
}

type cacheValue struct {
	atime time.Time
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
	value.atime = time.Now()

	if value.token == nil {
		return nil
	}

	res := copySessionTokenWithoutSignatureAndContext(*value.token)

	return &res
}

func (c *sessionCache) GetAccessTime(key string) (time.Time, bool) {
	valueRaw, ok := c.cache.Peek(key)
	if !ok {
		return time.Time{}, false
	}

	return valueRaw.(*cacheValue).atime, true
}

func (c *sessionCache) Put(key string, token *session.Token) bool {
	return c.cache.Add(key, &cacheValue{
		atime: time.Now(),
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
