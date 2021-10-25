package pool

import (
	"strings"

	"github.com/bluele/gcache"
	"github.com/nspcc-dev/neofs-api-go/pkg/session"
)

type SessionCache struct {
	cache gcache.Cache
}

func NewCache() *SessionCache {
	return &SessionCache{
		cache: gcache.New(100).Build(),
	}
}

func (c *SessionCache) Get(key string) *session.Token {
	tokenRaw, err := c.cache.Get(key)
	if err != nil {
		return nil
	}
	token, ok := tokenRaw.(*session.Token)
	if !ok {
		return nil
	}

	return token
}

func (c *SessionCache) Put(key string, token *session.Token) error {
	return c.cache.Set(key, token)
}

func (c *SessionCache) DeleteByPrefix(prefix string) {
	for _, key := range c.cache.Keys(false) {
		keyStr, ok := key.(string)
		if !ok {
			continue
		}
		if strings.HasPrefix(keyStr, prefix) {
			c.cache.Remove(key)
		}
	}
}
