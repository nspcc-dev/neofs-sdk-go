package container_test

import (
	"crypto/sha256"
	"strconv"
	"testing"
	"time"

	apicontainer "github.com/nspcc-dev/neofs-sdk-go/api/container"
	apinetmap "github.com/nspcc-dev/neofs-sdk-go/api/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	containertest "github.com/nspcc-dev/neofs-sdk-go/container/test"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestCalculateID(t *testing.T) {
	b := containertest.Container().Marshal()
	var c container.Container
	require.NoError(t, c.Unmarshal(b))
	require.EqualValues(t, sha256.Sum256(b), container.CalculateID(c))
}

func TestContainerDecoding(t *testing.T) {
	t.Run("missing fields", func(t *testing.T) {
		for _, testCase := range []struct {
			name, err string
			corrupt   func(*apicontainer.Container)
		}{
			{name: "version", err: "missing version", corrupt: func(c *apicontainer.Container) {
				c.Version = nil
			}},
			{name: "owner", err: "missing owner", corrupt: func(c *apicontainer.Container) {
				c.OwnerId = nil
			}},
			{name: "nil nonce", err: "missing nonce", corrupt: func(c *apicontainer.Container) {
				c.Nonce = nil
			}},
			{name: "empty nonce", err: "missing nonce", corrupt: func(c *apicontainer.Container) {
				c.Nonce = []byte{}
			}},
			{name: "policy", err: "missing placement policy", corrupt: func(c *apicontainer.Container) {
				c.PlacementPolicy = nil
			}},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				src := containertest.Container()
				var dst container.Container
				var m apicontainer.Container

				src.WriteToV2(&m)
				testCase.corrupt(&m)
				require.ErrorContains(t, dst.ReadFromV2(&m), testCase.err)

				b, err := proto.Marshal(&m)
				require.NoError(t, err)
				require.NoError(t, dst.Unmarshal(b))

				j, err := protojson.Marshal(&m)
				require.NoError(t, err)
				require.NoError(t, dst.UnmarshalJSON(j))
			})
		}
	})
	t.Run("invalid fields", func(t *testing.T) {
		for _, testCase := range []struct {
			name, err string
			corrupt   func(*apicontainer.Container)
		}{
			{name: "owner/nil value", err: "invalid owner: missing value", corrupt: func(c *apicontainer.Container) {
				c.OwnerId.Value = nil
			}},
			{name: "owner/empty value", err: "invalid owner: missing value", corrupt: func(c *apicontainer.Container) {
				c.OwnerId.Value = []byte{}
			}},
			{name: "owner/wrong length", err: "invalid owner: invalid value length 24", corrupt: func(c *apicontainer.Container) {
				c.OwnerId.Value = make([]byte, 24)
			}},
			{name: "owner/wrong prefix", err: "invalid owner: invalid prefix byte 0x34, expected 0x35", corrupt: func(c *apicontainer.Container) {
				c.OwnerId.Value[0] = 0x34
			}},
			{name: "owner/checksum mismatch", err: "invalid owner: value checksum mismatch", corrupt: func(c *apicontainer.Container) {
				c.OwnerId.Value[24]++
			}},
			{name: "nonce/wrong length", err: "invalid nonce: invalid UUID (got 15 bytes)", corrupt: func(c *apicontainer.Container) {
				c.Nonce = make([]byte, 15)
			}},
			{name: "nonce/wrong version", err: "invalid nonce: wrong UUID version 3", corrupt: func(c *apicontainer.Container) {
				c.Nonce[6] = 3 << 4
			}},
			{name: "nonce/nil replicas", err: "invalid placement policy: missing replicas", corrupt: func(c *apicontainer.Container) {
				c.PlacementPolicy.Replicas = nil
			}},
			{name: "attributes/empty key", err: "invalid attribute #1: missing key", corrupt: func(c *apicontainer.Container) {
				c.Attributes = []*apicontainer.Container_Attribute{
					{Key: "key_valid", Value: "any"},
					{Key: "", Value: "any"},
				}
			}},
			{name: "attributes/repeated keys", err: "multiple attributes with key=k2", corrupt: func(c *apicontainer.Container) {
				c.Attributes = []*apicontainer.Container_Attribute{
					{Key: "k1", Value: "any"},
					{Key: "k2", Value: "1"},
					{Key: "k3", Value: "any"},
					{Key: "k2", Value: "2"},
				}
			}},
			{name: "attributes/empty value", err: "invalid attribute #1 (key2): missing value", corrupt: func(c *apicontainer.Container) {
				c.Attributes = []*apicontainer.Container_Attribute{
					{Key: "key1", Value: "any"},
					{Key: "key2", Value: ""},
				}
			}},
			{name: "attributes/invalid timestamp", err: "invalid timestamp attribute (#1): invalid integer", corrupt: func(c *apicontainer.Container) {
				c.Attributes = []*apicontainer.Container_Attribute{
					{Key: "key1", Value: "any"},
					{Key: "Timestamp", Value: "not_a_number"},
				}
			}},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				src := containertest.Container()
				var dst container.Container
				var m apicontainer.Container

				src.WriteToV2(&m)
				testCase.corrupt(&m)
				require.ErrorContains(t, dst.ReadFromV2(&m), testCase.err)

				b, err := proto.Marshal(&m)
				require.NoError(t, err)
				require.ErrorContains(t, dst.Unmarshal(b), testCase.err)

				j, err := protojson.Marshal(&m)
				require.NoError(t, err)
				require.ErrorContains(t, dst.UnmarshalJSON(j), testCase.err)
			})
		}
	})
}

func TestContainer_Unmarshal(t *testing.T) {
	t.Run("invalid binary", func(t *testing.T) {
		var c container.Container
		msg := []byte("definitely_not_protobuf")
		err := c.Unmarshal(msg)
		require.ErrorContains(t, err, "decode protobuf")
	})
}

func TestNodeInfo_UnmarshalJSON(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		var c container.Container
		msg := []byte("definitely_not_protojson")
		err := c.UnmarshalJSON(msg)
		require.ErrorContains(t, err, "decode protojson")
	})
}

func TestContainer_CopyTo(t *testing.T) {
	src := containertest.Container()
	pp := src.PlacementPolicy()
	if len(pp.Replicas()) < 2 {
		pp.SetReplicas(netmaptest.NReplicas(2))
	}
	if len(pp.Filters()) < 2 {
		pp.SetFilters(netmaptest.NFilters(2))
	}
	if len(pp.Selectors()) < 2 {
		pp.SetSelectors(netmaptest.NSelectors(2))
	}
	src.SetPlacementPolicy(pp)
	const attr = "any_attr"
	src.SetAttribute(attr, "0")

	shallow := src
	deep := containertest.Container()
	src.CopyTo(&deep)
	require.Equal(t, src, deep)

	shallow.SetAttribute(attr, "1")
	require.Equal(t, "1", shallow.Attribute(attr))
	require.Equal(t, "1", src.Attribute(attr))

	deep.SetAttribute(attr, "2")
	require.Equal(t, "2", deep.Attribute(attr))
	require.Equal(t, "1", src.Attribute(attr))

	rs := src.PlacementPolicy().Replicas()
	originNum := rs[1].NumberOfObjects()
	rs[1].SetNumberOfObjects(originNum + 1)
	require.EqualValues(t, originNum+1, src.PlacementPolicy().Replicas()[1].NumberOfObjects())
	require.EqualValues(t, originNum+1, shallow.PlacementPolicy().Replicas()[1].NumberOfObjects())
	require.EqualValues(t, originNum, deep.PlacementPolicy().Replicas()[1].NumberOfObjects())

	fs := src.PlacementPolicy().Filters()
	originName := fs[1].Name()
	fs[1].SetName(originName + "_extra")
	require.EqualValues(t, originName+"_extra", src.PlacementPolicy().Filters()[1].Name())
	require.EqualValues(t, originName+"_extra", shallow.PlacementPolicy().Filters()[1].Name())
	require.EqualValues(t, originName, deep.PlacementPolicy().Filters()[1].Name())

	ss := src.PlacementPolicy().Selectors()
	originName = ss[1].Name()
	ss[1].SetName(originName + "_extra")
	require.EqualValues(t, originName+"_extra", src.PlacementPolicy().Selectors()[1].Name())
	require.EqualValues(t, originName+"_extra", shallow.PlacementPolicy().Selectors()[1].Name())
	require.EqualValues(t, originName, deep.PlacementPolicy().Selectors()[1].Name())
}

func assertPolicyAPIEncoding(t testing.TB, policy netmap.PlacementPolicy, msg *apinetmap.PlacementPolicy) {
	require.EqualValues(t, policy.ContainerBackupFactor(), msg.ContainerBackupFactor)

	if rs := policy.Replicas(); len(rs) > 0 {
		require.Len(t, msg.Replicas, len(rs))
		for i := range rs {
			require.EqualValues(t, rs[i].NumberOfObjects(), msg.Replicas[i].Count)
			require.Equal(t, rs[i].SelectorName(), msg.Replicas[i].Selector)
		}
	} else {
		require.Zero(t, msg.Replicas)
	}

	var assertFilters func(fs []netmap.Filter, m []*apinetmap.Filter)
	assertFilters = func(fs []netmap.Filter, m []*apinetmap.Filter) {
		if len(fs) > 0 {
			require.Len(t, m, len(fs))
			for i := range fs {
				require.Equal(t, fs[i].Name(), m[i].Name)
				require.Equal(t, fs[i].Key(), m[i].Key)
				require.EqualValues(t, fs[i].Op(), m[i].Op)
				require.Equal(t, fs[i].Value(), m[i].Value)
				assertFilters(fs[i].SubFilters(), m[i].Filters)
			}
		} else {
			require.Zero(t, m)
		}
	}

	assertFilters(policy.Filters(), msg.Filters)
	if ss := policy.Selectors(); len(ss) > 0 {
		require.Len(t, msg.Selectors, len(ss))
		for i := range ss {
			require.Equal(t, ss[i].Name(), msg.Selectors[i].Name)
			require.EqualValues(t, ss[i].NumberOfNodes(), msg.Selectors[i].Count)
			switch {
			default:
				require.Zero(t, msg.Selectors[i].Clause)
			case ss[i].IsSame():
				require.EqualValues(t, apinetmap.Clause_SAME, msg.Selectors[i].Clause)
			case ss[i].IsDistinct():
				require.EqualValues(t, apinetmap.Clause_DISTINCT, msg.Selectors[i].Clause)
			}
		}
	} else {
		require.Zero(t, msg.Selectors)
	}
}

func TestNew(t *testing.T) {
	owner := usertest.ID()
	basicACL := containertest.BasicACL()
	policy := netmaptest.PlacementPolicy()

	c := container.New(owner, basicACL, policy)
	require.Equal(t, owner, c.Owner())
	require.Equal(t, basicACL, c.BasicACL())
	require.Equal(t, policy, c.PlacementPolicy())

	assertFields := func(t testing.TB, cnr container.Container) {
		require.Empty(t, cnr.Name())
		require.Empty(t, cnr.Domain().Name())
		require.Zero(t, cnr.NumberOfAttributes())
		require.Zero(t, cnr.CreatedAt().Unix())
		require.False(t, cnr.HomomorphicHashingDisabled())
		require.EqualValues(t, 2, cnr.Version().Major())
		require.EqualValues(t, 16, cnr.Version().Minor())

		called := false
		f := func(key, val string) {
			called = true
		}
		cnr.IterateAttributes(f)
		require.False(t, called)
		cnr.IterateUserAttributes(f)
		require.False(t, called)

		netCfg := netmaptest.NetworkInfo()
		netCfg.SetHomomorphicHashingDisabled(true)
		require.False(t, cnr.AssertNetworkConfig(netCfg))
		netCfg.SetHomomorphicHashingDisabled(false)
		require.True(t, cnr.AssertNetworkConfig(netCfg))
	}
	assertFields(t, c)

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			src := container.New(owner, basicACL, policy)
			dst := containertest.Container()

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			assertFields(t, dst)
		})
		t.Run("api", func(t *testing.T) {
			src := container.New(owner, basicACL, policy)
			dst := containertest.Container()
			var msg apicontainer.Container

			src.WriteToV2(&msg)
			require.Equal(t, &refs.Version{Major: 2, Minor: 16}, msg.Version)
			require.Equal(t, &refs.OwnerID{Value: owner[:]}, msg.OwnerId)
			require.Len(t, msg.Nonce, 16)
			require.EqualValues(t, 4, msg.Nonce[6]>>4)
			require.EqualValues(t, basicACL, msg.BasicAcl)
			require.Zero(t, msg.Attributes)
			assertPolicyAPIEncoding(t, policy, msg.PlacementPolicy)
			require.NoError(t, dst.ReadFromV2(&msg))
			assertFields(t, dst)
		})
		t.Run("json", func(t *testing.T) {
			src := container.New(owner, basicACL, policy)
			dst := containertest.Container()

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			assertFields(t, dst)
		})
	})
}

func TestContainer_SetOwner(t *testing.T) {
	var c container.Container

	require.Zero(t, c.Owner())

	usr := usertest.ID()
	c.SetOwner(usr)
	require.Equal(t, usr, c.Owner())

	usrOther := usertest.ChangeID(usr)
	c.SetOwner(usrOther)
	require.Equal(t, usrOther, c.Owner())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst container.Container

			dst.SetOwner(usr)
			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Zero(t, dst.Owner())

			src.SetOwner(usr)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Equal(t, usr, dst.Owner())
		})
		t.Run("api", func(t *testing.T) {
			src := container.New(usr, containertest.BasicACL(), netmaptest.PlacementPolicy())
			var dst container.Container
			var msg apicontainer.Container

			dst.SetOwner(usrOther)
			src.SetOwner(usr)
			src.WriteToV2(&msg)
			require.Equal(t, &refs.OwnerID{Value: usr[:]}, msg.OwnerId)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Equal(t, usr, dst.Owner())
		})
		t.Run("json", func(t *testing.T) {
			var src, dst container.Container

			dst.SetOwner(usr)

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Zero(t, dst.Owner())

			src.SetOwner(usr)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Equal(t, usr, dst.Owner())
		})
	})
}

func TestContainer_SetBasicACL(t *testing.T) {
	var c container.Container

	require.Zero(t, c.BasicACL())

	basicACL := containertest.BasicACL()
	c.SetBasicACL(basicACL)
	require.Equal(t, basicACL, c.BasicACL())

	basicACLOther := containertest.BasicACL()
	c.SetBasicACL(basicACLOther)
	require.Equal(t, basicACLOther, c.BasicACL())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst container.Container

			dst.SetBasicACL(basicACL)
			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Zero(t, dst.BasicACL())

			src.SetBasicACL(basicACL)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Equal(t, basicACL, dst.BasicACL())
		})
		t.Run("api", func(t *testing.T) {
			src := container.New(usertest.ID(), basicACL, netmaptest.PlacementPolicy())
			var dst container.Container
			var msg apicontainer.Container

			src.SetBasicACL(basicACL)
			src.WriteToV2(&msg)
			require.EqualValues(t, basicACL, msg.BasicAcl)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Equal(t, basicACL, dst.BasicACL())
		})
		t.Run("json", func(t *testing.T) {
			var src, dst container.Container

			dst.SetBasicACL(basicACL)

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Zero(t, dst.BasicACL())

			src.SetBasicACL(basicACL)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Equal(t, basicACL, dst.BasicACL())
		})
	})
}

func TestContainer_SetPlacementPolicy(t *testing.T) {
	var c container.Container

	require.Zero(t, c.PlacementPolicy())

	policy := netmaptest.PlacementPolicy()
	c.SetPlacementPolicy(policy)
	require.Equal(t, policy, c.PlacementPolicy())

	policyOther := netmaptest.PlacementPolicy()
	c.SetPlacementPolicy(policyOther)
	require.Equal(t, policyOther, c.PlacementPolicy())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst container.Container

			dst.SetPlacementPolicy(policy)
			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Zero(t, dst.PlacementPolicy())

			src.SetPlacementPolicy(policy)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Equal(t, policy, dst.PlacementPolicy())
		})
		t.Run("api", func(t *testing.T) {
			src := container.New(usertest.ID(), containertest.BasicACL(), policy)
			var dst container.Container
			var msg apicontainer.Container

			src.SetPlacementPolicy(policy)
			src.WriteToV2(&msg)
			assertPolicyAPIEncoding(t, policy, msg.PlacementPolicy)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Equal(t, policy, dst.PlacementPolicy())
		})
		t.Run("json", func(t *testing.T) {
			var src, dst container.Container

			dst.SetPlacementPolicy(policy)

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Zero(t, dst.PlacementPolicy())

			src.SetPlacementPolicy(policy)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Equal(t, policy, dst.PlacementPolicy())
		})
	})
}

func collectContainerAttributes(c container.Container) [][2]string {
	var res [][2]string
	c.IterateAttributes(func(key, value string) {
		res = append(res, [2]string{key, value})
	})
	return res
}

func collectContainerUserAttributes(c container.Container) [][2]string {
	var res [][2]string
	c.IterateUserAttributes(func(key, value string) {
		res = append(res, [2]string{key, value})
	})
	return res
}

func TestContainer_SetAttribute(t *testing.T) {
	var c container.Container
	require.Panics(t, func() { c.SetAttribute("", "") })
	require.Panics(t, func() { c.SetAttribute("", "val") })
	require.Panics(t, func() { c.SetAttribute("key", "") })

	const key1, val1 = "some_key1", "some_value1"
	const key2, val2 = "some_key2", "some_value2"

	require.Zero(t, c.Attribute(key1))
	require.Zero(t, c.Attribute(key2))
	require.Zero(t, c.NumberOfAttributes())
	require.Zero(t, collectContainerAttributes(c))

	c.SetAttribute(key1, val1)
	c.SetAttribute(key2, val2)
	require.Equal(t, val1, c.Attribute(key1))
	require.Equal(t, val2, c.Attribute(key2))
	require.EqualValues(t, 2, c.NumberOfAttributes())
	attrs := collectContainerAttributes(c)
	require.Len(t, attrs, 2)
	require.Contains(t, attrs, [2]string{key1, val1})
	require.Contains(t, attrs, [2]string{key2, val2})

	c.SetAttribute(key1, val2)
	c.SetAttribute(key2, val1)
	require.Equal(t, val2, c.Attribute(key1))
	require.Equal(t, val1, c.Attribute(key2))
	attrs = collectContainerAttributes(c)
	require.Len(t, attrs, 2)
	require.Contains(t, attrs, [2]string{key1, val2})
	require.Contains(t, attrs, [2]string{key2, val1})

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst container.Container

			dst.SetAttribute(key1+key2, val1+val2)

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Zero(t, dst.Attribute(key1))
			require.Zero(t, dst.Attribute(key2))
			require.Zero(t, dst.NumberOfAttributes())
			require.Zero(t, collectContainerAttributes(dst))

			src.SetAttribute(key1, val1)
			src.SetAttribute(key2, val2)

			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Equal(t, val1, dst.Attribute(key1))
			require.Equal(t, val2, dst.Attribute(key2))
			require.EqualValues(t, 2, dst.NumberOfAttributes())
			attrs := collectContainerAttributes(dst)
			require.Len(t, attrs, 2)
			require.Contains(t, attrs, [2]string{key1, val1})
			require.Contains(t, attrs, [2]string{key2, val2})
		})
		t.Run("api", func(t *testing.T) {
			src := container.New(usertest.ID(), containertest.BasicACL(), netmaptest.PlacementPolicy())
			var dst container.Container
			var msg apicontainer.Container

			dst.SetAttribute(key1, val1)

			src.WriteToV2(&msg)
			require.Zero(t, msg.Attributes)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Zero(t, dst.Attribute(key1))
			require.Zero(t, dst.Attribute(key2))
			require.Zero(t, dst.NumberOfAttributes())
			require.Zero(t, collectContainerAttributes(dst))

			src.SetAttribute(key1, val1)
			src.SetAttribute(key2, val2)

			src.WriteToV2(&msg)
			require.Equal(t, []*apicontainer.Container_Attribute{
				{Key: key1, Value: val1},
				{Key: key2, Value: val2},
			}, msg.Attributes)

			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.Equal(t, val1, dst.Attribute(key1))
			require.Equal(t, val2, dst.Attribute(key2))
			require.EqualValues(t, 2, dst.NumberOfAttributes())
			attrs := collectContainerAttributes(dst)
			require.Len(t, attrs, 2)
			require.Contains(t, attrs, [2]string{key1, val1})
			require.Contains(t, attrs, [2]string{key2, val2})
		})
		t.Run("json", func(t *testing.T) {
			var src, dst container.Container

			dst.SetAttribute(key1, val1)

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Zero(t, dst.Attribute(key1))
			require.Zero(t, dst.Attribute(key2))
			require.Zero(t, dst.NumberOfAttributes())
			require.Zero(t, collectContainerAttributes(dst))

			src.SetAttribute(key1, val1)
			src.SetAttribute(key2, val2)

			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Equal(t, val1, dst.Attribute(key1))
			require.Equal(t, val2, dst.Attribute(key2))
			require.EqualValues(t, 2, dst.NumberOfAttributes())
			attrs := collectContainerAttributes(dst)
			require.Len(t, attrs, 2)
			require.Contains(t, attrs, [2]string{key1, val1})
			require.Contains(t, attrs, [2]string{key2, val2})
		})
	})
}

func TestContainer_SetName(t *testing.T) {
	var c container.Container

	assertZeroName := func(t testing.TB, c container.Container) {
		require.Zero(t, c.Name())
		require.Zero(t, c.Attribute("Name"))
		require.Zero(t, c.NumberOfAttributes())
		require.Zero(t, collectContainerAttributes(c))
		require.Zero(t, collectContainerUserAttributes(c))
	}
	assertName := func(t testing.TB, c container.Container, name string) {
		require.Equal(t, name, c.Name())
		require.Equal(t, name, c.Attribute("Name"))
		require.EqualValues(t, 1, c.NumberOfAttributes())
		require.Equal(t, [][2]string{{"Name", name}}, collectContainerAttributes(c))
		require.Equal(t, [][2]string{{"Name", name}}, collectContainerUserAttributes(c))
	}

	assertZeroName(t, c)

	name := "any_name"
	c.SetName(name)
	assertName(t, c, name)

	nameOther := name + "_extra"
	c.SetName(nameOther)
	assertName(t, c, nameOther)

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst container.Container

			dst.SetName(name)
			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			assertZeroName(t, dst)

			src.SetName(name)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			assertName(t, dst, name)
		})
		t.Run("api", func(t *testing.T) {
			src := container.New(usertest.ID(), containertest.BasicACL(), netmaptest.PlacementPolicy())
			var dst container.Container
			var msg apicontainer.Container

			dst.SetName(name)
			src.WriteToV2(&msg)
			require.Zero(t, msg.Attributes)
			require.NoError(t, dst.ReadFromV2(&msg))
			assertZeroName(t, dst)

			src.SetName(name)
			src.WriteToV2(&msg)
			require.Equal(t, []*apicontainer.Container_Attribute{
				{Key: "Name", Value: name},
			}, msg.Attributes)
			require.NoError(t, dst.ReadFromV2(&msg))
			assertName(t, dst, name)
		})
		t.Run("json", func(t *testing.T) {
			var src, dst container.Container

			dst.SetName(name)
			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			assertZeroName(t, dst)

			src.SetName(name)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			assertName(t, dst, name)
		})
	})
}

func TestContainer_SetCreationTime(t *testing.T) {
	var c container.Container

	assertZeroCreationTime := func(t testing.TB, c container.Container) {
		require.Zero(t, c.CreatedAt().Unix())
		require.Zero(t, c.Attribute("Timestamp"))
		require.Zero(t, c.NumberOfAttributes())
		require.Zero(t, collectContainerAttributes(c))
		require.Zero(t, collectContainerUserAttributes(c))
	}
	assertCreationTime := func(t testing.TB, c container.Container, tm time.Time) {
		require.Equal(t, tm.Unix(), c.CreatedAt().Unix())
		require.Equal(t, strconv.FormatInt(tm.Unix(), 10), c.Attribute("Timestamp"))
		require.EqualValues(t, 1, c.NumberOfAttributes())
		require.Equal(t, [][2]string{{"Timestamp", strconv.FormatInt(tm.Unix(), 10)}}, collectContainerAttributes(c))
		require.Equal(t, [][2]string{{"Timestamp", strconv.FormatInt(tm.Unix(), 10)}}, collectContainerUserAttributes(c))
	}

	tm := time.Now()
	c.SetCreationTime(tm)
	assertCreationTime(t, c, tm)

	tmOther := tm.Add(time.Minute)
	c.SetCreationTime(tmOther)
	assertCreationTime(t, c, tmOther)

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst container.Container

			dst.SetCreationTime(tm)
			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			assertZeroCreationTime(t, dst)

			src.SetCreationTime(tm)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			assertCreationTime(t, dst, tm)
		})
		t.Run("api", func(t *testing.T) {
			src := container.New(usertest.ID(), containertest.BasicACL(), netmaptest.PlacementPolicy())
			var dst container.Container
			var msg apicontainer.Container

			dst.SetCreationTime(tm)
			src.WriteToV2(&msg)
			require.Zero(t, msg.Attributes)
			require.NoError(t, dst.ReadFromV2(&msg))
			assertZeroCreationTime(t, dst)

			src.SetCreationTime(tm)
			src.WriteToV2(&msg)
			require.Equal(t, []*apicontainer.Container_Attribute{{
				Key: "Timestamp", Value: strconv.FormatInt(tm.Unix(), 10),
			}}, msg.Attributes)
			require.NoError(t, dst.ReadFromV2(&msg))
			assertCreationTime(t, dst, tm)
		})
		t.Run("json", func(t *testing.T) {
			var src, dst container.Container

			dst.SetCreationTime(tm)
			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			assertZeroCreationTime(t, dst)

			src.SetCreationTime(tm)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			assertCreationTime(t, dst, tm)
		})
	})
}

func assertHomomorphicHashingEnabled(t testing.TB, c container.Container, attr string) {
	require.False(t, c.HomomorphicHashingDisabled())
	require.Equal(t, attr, c.Attribute("__NEOFS__DISABLE_HOMOMORPHIC_HASHING"))
	require.Zero(t, collectContainerUserAttributes(c))
	if attr != "" {
		require.EqualValues(t, 1, c.NumberOfAttributes())
		require.Equal(t, [][2]string{{"__NEOFS__DISABLE_HOMOMORPHIC_HASHING", attr}}, collectContainerAttributes(c))
	} else {
		require.Zero(t, c.NumberOfAttributes())
		require.Zero(t, collectContainerAttributes(c))
	}
}

func assertHomomorphicHashingDisabled(t testing.TB, c container.Container) {
	require.True(t, c.HomomorphicHashingDisabled())
	require.Equal(t, "true", c.Attribute("__NEOFS__DISABLE_HOMOMORPHIC_HASHING"))
	require.EqualValues(t, 1, c.NumberOfAttributes())
	require.Equal(t, [][2]string{{"__NEOFS__DISABLE_HOMOMORPHIC_HASHING", "true"}}, collectContainerAttributes(c))
	require.Zero(t, collectContainerUserAttributes(c))
}

func TestContainer_SetHomomorphicHashingDisabled(t *testing.T) {
	var c container.Container

	assertHomomorphicHashingEnabled(t, c, "")

	c.SetHomomorphicHashingDisabled(true)
	assertHomomorphicHashingDisabled(t, c)

	c.SetHomomorphicHashingDisabled(false)
	assertHomomorphicHashingEnabled(t, c, "")

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst container.Container

			dst.SetHomomorphicHashingDisabled(true)
			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			assertHomomorphicHashingEnabled(t, dst, "")

			src.SetHomomorphicHashingDisabled(true)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			assertHomomorphicHashingDisabled(t, dst)

			src.SetAttribute("__NEOFS__DISABLE_HOMOMORPHIC_HASHING", "any")
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			assertHomomorphicHashingEnabled(t, dst, "any")
		})
		t.Run("api", func(t *testing.T) {
			src := container.New(usertest.ID(), containertest.BasicACL(), netmaptest.PlacementPolicy())
			var dst container.Container
			var msg apicontainer.Container

			dst.SetHomomorphicHashingDisabled(true)
			src.WriteToV2(&msg)
			require.Zero(t, msg.Attributes)
			require.NoError(t, dst.ReadFromV2(&msg))
			assertHomomorphicHashingEnabled(t, dst, "")

			src.SetHomomorphicHashingDisabled(true)
			src.WriteToV2(&msg)
			require.Equal(t, []*apicontainer.Container_Attribute{{
				Key: "__NEOFS__DISABLE_HOMOMORPHIC_HASHING", Value: "true",
			}}, msg.Attributes)
			require.NoError(t, dst.ReadFromV2(&msg))
			assertHomomorphicHashingDisabled(t, dst)

			msg.Attributes[0].Value = "any"
			require.NoError(t, dst.ReadFromV2(&msg))
			assertHomomorphicHashingEnabled(t, dst, "any")
		})
		t.Run("json", func(t *testing.T) {
			var src, dst container.Container

			dst.SetHomomorphicHashingDisabled(true)
			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			assertHomomorphicHashingEnabled(t, dst, "")

			src.SetHomomorphicHashingDisabled(true)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			assertHomomorphicHashingDisabled(t, dst)

			src.SetAttribute("__NEOFS__DISABLE_HOMOMORPHIC_HASHING", "any")
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			assertHomomorphicHashingEnabled(t, dst, "any")
		})
	})
}

func TestContainer_SetDomain(t *testing.T) {
	var c container.Container

	assertNoDomain := func(t testing.TB, c container.Container) {
		require.Zero(t, c.Domain())
		require.Zero(t, c.NumberOfAttributes())
		require.Zero(t, c.Attribute("__NEOFS__NAME"))
		require.Zero(t, c.Attribute("__NEOFS__ZONE"))
		require.Zero(t, collectContainerAttributes(c))
		require.Zero(t, collectContainerUserAttributes(c))
	}
	assertDomain := func(t testing.TB, c container.Container, name, zone string) {
		require.Equal(t, name, c.Domain().Name())
		require.Equal(t, name, c.Attribute("__NEOFS__NAME"))
		require.Zero(t, collectContainerUserAttributes(c))
		if zone != "" && zone != "container" {
			require.Equal(t, zone, c.Domain().Zone())
			require.Equal(t, zone, c.Attribute("__NEOFS__ZONE"))
			require.EqualValues(t, 2, c.NumberOfAttributes())
			require.ElementsMatch(t, [][2]string{
				{"__NEOFS__NAME", name},
				{"__NEOFS__ZONE", zone},
			}, collectContainerAttributes(c))
		} else {
			require.Equal(t, "container", c.Domain().Zone())
			require.Zero(t, c.Attribute("__NEOFS__ZONE"))
			require.EqualValues(t, 1, c.NumberOfAttributes())
			require.Equal(t, [][2]string{{"__NEOFS__NAME", name}}, collectContainerAttributes(c))
		}
	}

	assertNoDomain(t, c)

	var domain container.Domain
	domain.SetName("name")

	c.SetDomain(domain)
	assertDomain(t, c, "name", "")

	var domainOther container.Domain
	domainOther.SetName("name_other")
	domainOther.SetZone("zone")
	c.SetDomain(domainOther)
	assertDomain(t, c, "name_other", "zone")

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst container.Container

			dst.SetDomain(domain)
			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			assertNoDomain(t, dst)

			src.SetDomain(domain)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			assertDomain(t, dst, "name", "")

			src.SetDomain(domainOther)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			assertDomain(t, dst, "name_other", "zone")
		})
		t.Run("api", func(t *testing.T) {
			src := container.New(usertest.ID(), containertest.BasicACL(), netmaptest.PlacementPolicy())
			var dst container.Container
			var msg apicontainer.Container

			dst.SetDomain(domain)
			src.WriteToV2(&msg)
			require.Zero(t, msg.Attributes)
			require.NoError(t, dst.ReadFromV2(&msg))
			assertNoDomain(t, dst)

			src.SetDomain(domain)
			src.WriteToV2(&msg)
			require.Equal(t, []*apicontainer.Container_Attribute{{
				Key: "__NEOFS__NAME", Value: "name",
			}}, msg.Attributes)
			require.NoError(t, dst.ReadFromV2(&msg))
			assertDomain(t, dst, "name", "")

			src.SetDomain(domainOther)
			src.WriteToV2(&msg)
			require.ElementsMatch(t, []*apicontainer.Container_Attribute{
				{Key: "__NEOFS__NAME", Value: "name_other"},
				{Key: "__NEOFS__ZONE", Value: "zone"},
			}, msg.Attributes)
			require.NoError(t, dst.ReadFromV2(&msg))
			assertDomain(t, dst, "name_other", "zone")
		})
		t.Run("json", func(t *testing.T) {
			var src, dst container.Container

			dst.SetDomain(domain)
			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			assertNoDomain(t, dst)

			src.SetDomain(domain)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			assertDomain(t, dst, "name", "")

			src.SetDomain(domainOther)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			assertDomain(t, dst, "name_other", "zone")
		})
	})
}
