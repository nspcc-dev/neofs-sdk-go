package policy

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/stretchr/testify/require"
)

func TestToJSON(t *testing.T) {
	check := func(t *testing.T, p *netmap.PlacementPolicy, json string) {
		data, err := ToJSON(p)
		require.NoError(t, err)
		require.JSONEq(t, json, string(data))

		np, err := FromJSON(data)
		require.NoError(t, err)
		require.Equal(t, p, np)
	}
	t.Run("SimpleREP", func(t *testing.T) {
		p := new(netmap.PlacementPolicy)
		p.SetReplicas(newReplica("", 3))
		check(t, p, `{"replicas":[{"count":3}]}`)
	})
	t.Run("REPWithCBF", func(t *testing.T) {
		p := new(netmap.PlacementPolicy)
		p.SetReplicas(newReplica("", 3))
		p.SetContainerBackupFactor(3)
		check(t, p, `{"replicas":[{"count":3}],"container_backup_factor":3}`)
	})
	t.Run("REPFromSelector", func(t *testing.T) {
		p := new(netmap.PlacementPolicy)
		p.SetReplicas(newReplica("Nodes", 3))
		p.SetContainerBackupFactor(3)
		p.SetSelectors(
			newSelector(1, netmap.ClauseDistinct, "City", "", "Nodes"))
		check(t, p, `{
			"replicas":[{"count":3,"selector":"Nodes"}],
			"container_backup_factor":3,
			"selectors": [{
				"name":"Nodes",
				"attribute":"City",
				"clause":"distinct",
				"count":1
			}]}`)
	})
	t.Run("FilterOps", func(t *testing.T) {
		p := new(netmap.PlacementPolicy)
		p.SetReplicas(newReplica("Nodes", 3))
		p.SetContainerBackupFactor(3)
		p.SetSelectors(
			newSelector(1, netmap.ClauseSame, "City", "Good", "Nodes"))
		p.SetFilters(
			newFilter("GoodRating", "Rating", "5", netmap.OpGE),
			newFilter("Good", "", "", netmap.OpOR,
				newFilter("GoodRating", "", "", 0),
				newFilter("", "Attr1", "Val1", netmap.OpEQ),
				newFilter("", "Attr2", "Val2", netmap.OpNE),
				newFilter("", "", "", netmap.OpAND,
					newFilter("", "Attr4", "2", netmap.OpLT),
					newFilter("", "Attr5", "3", netmap.OpLE)),
				newFilter("", "Attr3", "1", netmap.OpGT)),
		)
		check(t, p, `{
			"replicas":[{"count":3,"selector":"Nodes"}],
			"container_backup_factor":3,
			"selectors": [{"name":"Nodes","attribute":"City","clause":"same","count":1,"filter":"Good"}],
			"filters": [
				{"name":"GoodRating","key":"Rating","op":"GE","value":"5"},
				{"name":"Good","op":"OR","filters":[
					{"name":"GoodRating"},
					{"key":"Attr1","op":"EQ","value":"Val1"},
					{"key":"Attr2","op":"NE","value":"Val2"},
					{"op":"AND","filters":[
						{"key":"Attr4","op":"LT","value":"2"},
						{"key":"Attr5","op":"LE","value":"3"}
					]},
					{"key":"Attr3","op":"GT","value":"1"}
				]}
			]}`)
	})
}
