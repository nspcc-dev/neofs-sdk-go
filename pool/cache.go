package pool

import (
	"strings"
	"sync/atomic"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

type SessionCache struct {
	cache        *lru.Cache
	currentEpoch uint64
}

type cacheValue struct {
	atime time.Time
	token *session.Token
}

func NewCache() (*SessionCache, error) {
	cache, err := lru.New(100)
	if err != nil {
		return nil, err
	}

	return &SessionCache{cache: cache}, nil
}

func (c *SessionCache) Get(key string) *session.Token {
	valueRaw, ok := c.cache.Get(key)
	if !ok {
		return nil
	}

	value := valueRaw.(*cacheValue)
	if c.expired(value) {
		c.cache.Remove(key)
		return nil
	}

	value.atime = time.Now()

	return value.token
}

func (c *SessionCache) GetAccessTime(key string) (time.Time, bool) {
	valueRaw, ok := c.cache.Peek(key)
	if !ok {
		return time.Time{}, false
	}

	return valueRaw.(*cacheValue).atime, true
}

func (c *SessionCache) Put(key string, token *session.Token) bool {
	return c.cache.Add(key, &cacheValue{
		atime: time.Now(),
		token: token,
	})
}

func (c *SessionCache) DeleteByPrefix(prefix string) {
	for _, key := range c.cache.Keys() {
		if strings.HasPrefix(key.(string), prefix) {
			c.cache.Remove(key)
		}
	}
}

func (c *SessionCache) UpdateEpoch(epoch uint64) {
	atomic.StoreUint64(&c.currentEpoch, epoch)
}

func (c *SessionCache) expired(val *cacheValue) bool {
	epoch := atomic.LoadUint64(&c.currentEpoch)
	return val.token.Exp() <= epoch
}
