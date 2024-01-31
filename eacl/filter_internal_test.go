package eacl

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFilter(t *testing.T) {
	const anyValidType = HeaderFromRequest
	const anyValidKey = "any_key"
	const anyValidMatcher = MatchStringNotEqual
	const anyValidVal = "any_value"

	for _, typ := range []HeaderType{
		0,
		lastHeaderType,
		lastHeaderType + 1,
	} {
		t.Run(fmt.Sprintf("invalid type=%d", typ), func(t *testing.T) {
			require.Panics(t, func() { NewFilter(typ, anyValidKey, anyValidMatcher, anyValidVal) })
		})
	}

	for _, matcher := range []Matcher{
		0,
		lastMatcher,
		lastMatcher + 1,
	} {
		t.Run(fmt.Sprintf("invalid matcher=%d", matcher), func(t *testing.T) {
			require.Panics(t, func() { NewFilter(anyValidType, anyValidKey, matcher, anyValidVal) })
		})
	}
}

func TestFilterDeepCopy(t *testing.T) {
	const srcType = HeaderFromRequest
	const srcKey = "any_key"
	const srcMatcher = MatchStringNotEqual
	const srcVal = "any_value"

	src := NewFilter(srcType, srcKey, srcMatcher, srcVal)

	var dst Filter
	src.copyTo(&dst)

	require.Equal(t, src, dst)

	changeAllFilterFields(&dst)

	require.Equal(t, srcType, src.hdrType)
	require.Equal(t, srcMatcher, src.matcher)
	require.Equal(t, srcKey, src.key)
	require.Equal(t, srcVal, src.value)

	t.Run("full-to-full", func(t *testing.T) {
		dst = NewFilter(HeaderFromObject, "other_key", MatchStringEqual, "other_val")
		src.copyTo(&dst)
		require.Equal(t, src, dst)
	})

	t.Run("zero-to-full", func(t *testing.T) {
		var zero Filter
		zero.copyTo(&src)
		require.Zero(t, src)
	})
}
