package netmap

import (
	"math/rand"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	testCases := []string{
		`REP 1 IN X
CBF 1
SELECT 2 IN SAME Location FROM * AS X`,

		`REP 1
SELECT 2 IN City FROM Good
FILTER Country EQ RU AS FromRU
FILTER @FromRU AND Rating GT 7 AS Good`,

		`REP 7 IN SPB
SELECT 1 IN City FROM SPBSSD AS SPB
FILTER City EQ SPB AND SSD EQ true OR City EQ SPB AND Rating GE 5 AS SPBSSD`,
	}

	for _, testCase := range testCases {
		var p PlacementPolicy

		require.NoError(t, p.DecodeString(testCase))

		var b strings.Builder
		require.NoError(t, p.WriteStringTo(&b))

		require.Equal(t, testCase, b.String())
	}
}

type cache struct {
	mtx sync.RWMutex

	item map[string]struct{}
}

func (x *cache) add(key string) {
	x.mtx.Lock()
	x.item[key] = struct{}{}
	x.mtx.Unlock()
}

func (x *cache) has(key string) bool {
	x.mtx.RLock()
	_, ok := x.item[key]
	x.mtx.RUnlock()

	return ok
}

func BenchmarkCache(b *testing.B) {
	c := cache{
		item: make(map[string]struct{}),
	}

	var key string
	buf := make([]byte, 32)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rand.Read(buf)
		key = string(buf)
		b.StartTimer()

		c.add(key)
		c.has(key)
	}
}

type cacheP struct {
	mtx *sync.RWMutex

	item map[string]struct{}
}

func (x cacheP) add(key string) {
	x.mtx.Lock()
	x.item[key] = struct{}{}
	x.mtx.Unlock()
}

func (x cacheP) has(key string) bool {
	x.mtx.RLock()
	_, ok := x.item[key]
	x.mtx.RUnlock()

	return ok
}

func BenchmarkCacheP(b *testing.B) {
	c := cacheP{
		mtx:  &sync.RWMutex{},
		item: make(map[string]struct{}),
	}

	var key string
	buf := make([]byte, 32)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rand.Read(buf)
		key = string(buf)
		b.StartTimer()

		c.add(key)
		c.has(key)
	}
}
