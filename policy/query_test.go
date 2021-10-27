package policy

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/stretchr/testify/require"
)

func TestSimple(t *testing.T) {
	q := `REP 3`
	expected := new(netmap.PlacementPolicy)
	expected.SetFilters([]*netmap.Filter{}...)
	expected.SetSelectors([]*netmap.Selector{}...)
	expected.SetReplicas(newReplica("", 3))

	r, err := Parse(q)
	require.NoError(t, err)
	require.EqualValues(t, expected, r)
}

func TestSimpleWithHRWB(t *testing.T) {
	q := `REP 3 CBF 4`
	expected := new(netmap.PlacementPolicy)
	expected.SetFilters([]*netmap.Filter{}...)
	expected.SetSelectors([]*netmap.Selector{}...)
	expected.SetReplicas(newReplica("", 3))
	expected.SetContainerBackupFactor(4)

	r, err := Parse(q)
	require.NoError(t, err)
	require.EqualValues(t, expected, r)
}

func TestFromSelect(t *testing.T) {
	q := `REP 1 IN SPB
SELECT 1 IN City FROM * AS SPB`
	expected := new(netmap.PlacementPolicy)
	expected.SetFilters([]*netmap.Filter{}...)
	expected.SetSelectors(newSelector(1, netmap.ClauseUnspecified, "City", "*", "SPB"))
	expected.SetReplicas(newReplica("SPB", 1))

	r, err := Parse(q)
	require.NoError(t, err)
	require.EqualValues(t, expected, r)
}

// https://github.com/nspcc-dev/neofs-node/issues/46
func TestFromSelectNoAttribute(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		q := `REP 2
		SELECT 6 FROM *`

		expected := new(netmap.PlacementPolicy)
		expected.SetFilters([]*netmap.Filter{}...)
		expected.SetSelectors(newSelector(6, netmap.ClauseUnspecified, "", "*", ""))
		expected.SetReplicas(newReplica("", 2))

		r, err := Parse(q)
		require.NoError(t, err)
		require.EqualValues(t, expected, r)
	})

	t.Run("with filter", func(t *testing.T) {
		q := `REP 2
		SELECT 6 FROM F
		FILTER StorageType EQ SSD AS F`

		expected := new(netmap.PlacementPolicy)
		expected.SetFilters(newFilter("F", "StorageType", "SSD", netmap.OpEQ))
		expected.SetSelectors(newSelector(6, netmap.ClauseUnspecified, "", "F", ""))
		expected.SetReplicas(newReplica("", 2))

		r, err := Parse(q)
		require.NoError(t, err)
		require.EqualValues(t, expected, r)
	})
}

func TestString(t *testing.T) {
	qTemplate := `REP 1
SELECT 1 IN City FROM Filt
FILTER Property EQ %s AND Something NE 7 AS Filt`

	testCases := []string{
		`"double-quoted"`,
		`"with ' single"`,
		`'single-quoted'`,
		`'with " double'`,
	}

	for _, s := range testCases {
		t.Run(s, func(t *testing.T) {
			q := fmt.Sprintf(qTemplate, s)
			r, err := Parse(q)
			require.NoError(t, err)

			expected := newFilter("Filt", "", "", netmap.OpAND,
				newFilter("", "Property", s[1:len(s)-1], netmap.OpEQ),
				newFilter("", "Something", "7", netmap.OpNE))
			require.EqualValues(t, []*netmap.Filter{expected}, r.Filters())
		})
	}
}

func TestFromSelectClause(t *testing.T) {
	q := `REP 4
SELECT 3 IN Country FROM *
SELECT 2 IN SAME City FROM *
SELECT 1 IN DISTINCT Continent FROM *`
	expected := new(netmap.PlacementPolicy)
	expected.SetFilters([]*netmap.Filter{}...)
	expected.SetSelectors(
		newSelector(3, netmap.ClauseUnspecified, "Country", "*", ""),
		newSelector(2, netmap.ClauseSame, "City", "*", ""),
		newSelector(1, netmap.ClauseDistinct, "Continent", "*", ""))
	expected.SetReplicas(newReplica("", 4))

	r, err := Parse(q)
	require.NoError(t, err)
	require.EqualValues(t, expected, r)
}

func TestSimpleFilter(t *testing.T) {
	q := `REP 1
SELECT 1 IN City FROM Good
FILTER Rating GT 7 AS Good`
	expected := new(netmap.PlacementPolicy)
	expected.SetReplicas(newReplica("", 1))
	expected.SetSelectors(
		newSelector(1, netmap.ClauseUnspecified, "City", "Good", ""))
	expected.SetFilters(newFilter("Good", "Rating", "7", netmap.OpGT))

	r, err := Parse(q)
	require.NoError(t, err)
	require.EqualValues(t, expected, r)
}

func TestFilterReference(t *testing.T) {
	q := `REP 1
SELECT 2 IN City FROM Good
FILTER Country EQ "RU" AS FromRU
FILTER @FromRU AND Rating GT 7 AS Good`
	expected := new(netmap.PlacementPolicy)
	expected.SetReplicas(newReplica("", 1))
	expected.SetSelectors(
		newSelector(2, netmap.ClauseUnspecified, "City", "Good", ""))
	expected.SetFilters(
		newFilter("FromRU", "Country", "RU", netmap.OpEQ),
		newFilter("Good", "", "", netmap.OpAND,
			newFilter("FromRU", "", "", 0),
			newFilter("", "Rating", "7", netmap.OpGT)),
	)

	r, err := Parse(q)
	require.NoError(t, err)
	require.EqualValues(t, expected, r)
}

func TestFilterOps(t *testing.T) {
	q := `REP 1
SELECT 2 IN City FROM Good
FILTER A GT 1 AND B GE 2 AND C LT 3 AND D LE 4
  AND E EQ 5 AND F NE 6 AS Good`
	expected := new(netmap.PlacementPolicy)
	expected.SetReplicas(newReplica("", 1))
	expected.SetSelectors(
		newSelector(2, netmap.ClauseUnspecified, "City", "Good", ""))
	expected.SetFilters(newFilter("Good", "", "", netmap.OpAND,
		newFilter("", "A", "1", netmap.OpGT),
		newFilter("", "B", "2", netmap.OpGE),
		newFilter("", "C", "3", netmap.OpLT),
		newFilter("", "D", "4", netmap.OpLE),
		newFilter("", "E", "5", netmap.OpEQ),
		newFilter("", "F", "6", netmap.OpNE)))

	r, err := Parse(q)
	require.NoError(t, err)
	require.EqualValues(t, expected, r)
}

func TestWithFilterPrecedence(t *testing.T) {
	q := `REP 7 IN SPB
SELECT 1 IN City FROM SPBSSD AS SPB
FILTER City EQ "SPB" AND SSD EQ true OR City EQ "SPB" AND Rating GE 5 AS SPBSSD`
	expected := new(netmap.PlacementPolicy)
	expected.SetReplicas(newReplica("SPB", 7))
	expected.SetSelectors(
		newSelector(1, netmap.ClauseUnspecified, "City", "SPBSSD", "SPB"))
	expected.SetFilters(
		newFilter("SPBSSD", "", "", netmap.OpOR,
			newFilter("", "", "", netmap.OpAND,
				newFilter("", "City", "SPB", netmap.OpEQ),
				newFilter("", "SSD", "true", netmap.OpEQ)),
			newFilter("", "", "", netmap.OpAND,
				newFilter("", "City", "SPB", netmap.OpEQ),
				newFilter("", "Rating", "5", netmap.OpGE))))

	r, err := Parse(q)
	require.NoError(t, err)
	require.EqualValues(t, expected, r)
}

func TestBrackets(t *testing.T) {
	q := `REP 7 IN SPB
SELECT 1 IN City FROM SPBSSD AS SPB
FILTER ( City EQ "SPB" OR SSD EQ true ) AND (City EQ "SPB" OR Rating GE 5) AS SPBSSD`

	expected := new(netmap.PlacementPolicy)
	expected.SetReplicas(newReplica("SPB", 7))
	expected.SetSelectors(
		newSelector(1, netmap.ClauseUnspecified, "City", "SPBSSD", "SPB"))
	expected.SetFilters(
		newFilter("SPBSSD", "", "", netmap.OpAND,
			newFilter("", "", "", netmap.OpOR,
				newFilter("", "City", "SPB", netmap.OpEQ),
				newFilter("", "SSD", "true", netmap.OpEQ)),
			newFilter("", "", "", netmap.OpOR,
				newFilter("", "City", "SPB", netmap.OpEQ),
				newFilter("", "Rating", "5", netmap.OpGE))))

	r, err := Parse(q)
	require.NoError(t, err)
	require.EqualValues(t, expected, r)
}

func TestValidation(t *testing.T) {
	t.Run("MissingSelector", func(t *testing.T) {
		q := `REP 3 IN RU`
		_, err := Parse(q)
		require.True(t, errors.Is(err, ErrUnknownSelector), "got: %v", err)
	})
	t.Run("MissingFilter", func(t *testing.T) {
		q := `REP 3
              SELECT 1 IN City FROM MissingFilter`
		_, err := Parse(q)
		require.True(t, errors.Is(err, ErrUnknownFilter), "got: %v", err)
	})
	t.Run("UnknownOp", func(t *testing.T) {
		q := `REP 3
              SELECT 1 IN City FROM F
			  FILTER Country KEK RU AS F`
		_, err := Parse(q)
		require.True(t, errors.Is(err, ErrSyntaxError), "got: %v", err)
	})
	t.Run("TypoInREP", func(t *testing.T) {
		q := `REK 3`
		_, err := Parse(q)
		require.True(t, errors.Is(err, ErrSyntaxError))
	})
	t.Run("InvalidFilterName", func(t *testing.T) {
		q := `REP 3
			  SELECT 1 IN City FROM F
			  FILTER Good AND Country EQ RU AS F
			  FILTER Rating EQ 5 AS Good`
		_, err := Parse(q)
		require.Error(t, err)
	})
}

// Checks that an error is returned in cases when positive 32-bit integer is expected.
func TestInvalidNumbers(t *testing.T) {
	tmpls := []string{
		"REP %d",
		"REP 1 CBF %d",
		"REP 1 SELECT %d FROM *",
	}
	for i := range tmpls {
		zero := fmt.Sprintf(tmpls[i], 0)
		t.Run(zero, func(t *testing.T) {
			_, err := Parse(zero)
			require.Error(t, err)
		})

		big := fmt.Sprintf(tmpls[i], int64(math.MaxUint32)+1)
		t.Run(big, func(t *testing.T) {
			_, err := Parse(big)
			require.Error(t, err)
		})
	}
}

func TestFilterStringSymbols(t *testing.T) {
	q := `REP 1 IN S
SELECT 1 FROM F AS S
FILTER "UN-LOCODE" EQ "RU LED" AS F`

	expected := new(netmap.PlacementPolicy)
	expected.SetReplicas(newReplica("S", 1))
	expected.SetSelectors(
		newSelector(1, netmap.ClauseUnspecified, "", "F", "S"))
	expected.SetFilters(newFilter("F", "UN-LOCODE", "RU LED", netmap.OpEQ))

	r, err := Parse(q)
	require.NoError(t, err)
	require.EqualValues(t, expected, r)
}

func newFilter(name, key, value string, op netmap.Operation, sub ...*netmap.Filter) *netmap.Filter {
	f := new(netmap.Filter)
	f.SetName(name)
	f.SetKey(key)
	f.SetValue(value)
	f.SetOperation(op)
	f.SetInnerFilters(sub...)
	return f
}

func newReplica(s string, c uint32) *netmap.Replica {
	r := new(netmap.Replica)
	r.SetSelector(s)
	r.SetCount(c)
	return r
}

func newSelector(count uint32, c netmap.Clause, attr, f, name string) *netmap.Selector {
	s := new(netmap.Selector)
	s.SetCount(count)
	s.SetClause(c)
	s.SetAttribute(attr)
	s.SetFilter(f)
	s.SetName(name)
	return s
}
