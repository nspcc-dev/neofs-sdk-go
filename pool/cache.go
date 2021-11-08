package pool

import (
	"strings"

	lru "github.com/hashicorp/golang-lru"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

type SessionCache struct {
	cache *lru.Cache
}

func NewCache() (*SessionCache, error) {
	cache, err := lru.New(100)
	if err != nil {
		return nil, err
	}

	return &SessionCache{cache: cache}, nil
}

func (c *SessionCache) Get(key string) *session.Token {
	tokenRaw, ok := c.cache.Get(key)
	if !ok {
		return nil
	}
	return tokenRaw.(*session.Token)
}

func (c *SessionCache) Put(key string, token *session.Token) bool {
	return c.cache.Add(key, token)
}

func (c *SessionCache) DeleteByPrefix(prefix string) {
	for _, key := range c.cache.Keys() {
		if strings.HasPrefix(key.(string), prefix) {
			c.cache.Remove(key)
		}
	}
}
