package netmap

import (
	"strings"
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
