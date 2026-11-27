//go:build aiotest

package pool_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/internal/testutil"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/pool"
	sessionv2 "github.com/nspcc-dev/neofs-sdk-go/session/v2"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	defaultTimeOut = 5 * time.Second
	tickInterval   = 1 * time.Second

	tickEpochCmdBuilder = func(epoch uint64) []string {
		return []string{
			"neo-go", "contract", "invokefunction", "--wallet-config", "/config/node-config.yaml",
			"-a", "NfgHwwTi3wHAS8aFAN243C5vGbkYDpqLHP", "--force", "-r", "http://localhost:30333",
			"707516630852f4179af43366917a36b9a78b93a5", "newEpoch", fmt.Sprintf("int:%d", epoch),
			"--", "NfgHwwTi3wHAS8aFAN243C5vGbkYDpqLHP:Global",
		}
	}

	tickNewEpoch newEpochTickerFunc

	versions = []dockerImage{
		{image: "nspccdev/neofs-aio", version: "0.51.1"},
		{image: "nspccdev/neofs-aio", version: "latest"},
	}

	sessionExpirationInEpochs  = uint64(2)
	sessionExpirationInSeconds = uint64(3600) // 1 hour

	// clientRebalanceInterval must be lower than timeoutAfterEpochChange.
	// It is important if you are forcing epoch changing. Otherwise, session tokens maybe expired and rejected by node.
	// It happens because session cache inside pool just wasn't updated yet.
	clientRebalanceInterval = 1 * time.Second
	timeoutAfterEpochChange = 2 * time.Second
)

type (
	NetworkInfoExecutor interface {
		NetworkInfo(ctx context.Context, prm client.PrmNetworkInfo) (netmap.NetworkInfo, error)
	}

	newEpochTickerFunc func(context.Context, NetworkInfoExecutor) (int, error)

	dockerImage struct {
		image   string
		version string
	}

	containerCreator interface {
		ContainerPut(ctx context.Context, cont container.Container, signer neofscrypto.Signer, prm client.PrmContainerPut) (cid.ID, error)
	}

	containerDeleter interface {
		ContainerDelete(ctx context.Context, id cid.ID, signer neofscrypto.Signer, prm client.PrmContainerDelete) error
	}

	objectPutIniter interface {
		ObjectPutInit(ctx context.Context, hdr object.Object, signer user.Signer, prm client.PrmObjectPutInit) (client.ObjectWriter, error)
	}

	objectDeleter interface {
		ObjectDelete(ctx context.Context, containerID cid.ID, objectID oid.ID, signer user.Signer, prm client.PrmObjectDelete) (oid.ID, error)
	}

	containerEaclSetter interface {
		ContainerSetEACL(ctx context.Context, table eacl.Table, signer user.Signer, prm client.PrmContainerSetEACL) error
	}

	containerEaclGetter interface {
		ContainerEACL(ctx context.Context, id cid.ID, prm client.PrmContainerEACL) (eacl.Table, error)
	}

	containerGetter interface {
		ContainerGet(ctx context.Context, id cid.ID, prm client.PrmContainerGet) (container.Container, error)
	}

	objectHeadGetter interface {
		ObjectHead(ctx context.Context, containerID cid.ID, objectID oid.ID, signer user.Signer, prm client.PrmObjectHead) (*object.Object, error)
	}
)

func testContainer(account user.ID, basicACL acl.Basic) container.Container {
	containerName := strconv.FormatInt(time.Now().UnixNano(), 16)
	creationTime := time.Now()

	var cont container.Container
	cont.Init()
	cont.SetBasicACL(basicACL)
	cont.SetOwner(account)
	cont.SetName(containerName)
	cont.SetCreationTime(creationTime)

	return cont
}

func testData(_ *testing.T) (user.ID, user.Signer, container.Container) {
	signer := usertest.User().RFC6979
	account := signer.UserID()

	return account, signer, testContainer(account, acl.PublicRWExtended)
}

func testEaclOthersTable(containerID cid.ID, op eacl.Operation, action eacl.Action) eacl.Table {
	var r eacl.Record
	r.SetOperation(op)
	r.SetAction(action)

	var target eacl.Target
	target.SetRole(eacl.RoleOthers)
	r.SetTargets(target)

	return eacl.NewTableForContainer(containerID, []eacl.Record{r})
}

func testEaclTable(containerID cid.ID) eacl.Table {
	return testEaclOthersTable(containerID, eacl.OperationPut, eacl.ActionAllow)
}

func TestPoolAio(t *testing.T) {
	for _, version := range versions {
		image := fmt.Sprintf("%s:%s", version.image, version.version)

		t.Run(image, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			aioContainer := createDockerContainer(ctx, t, image)

			// for instance: localhost:32781
			nodeEndpoint, err := aioContainer.Endpoint(ctx, "")
			require.NoError(t, err)

			runTests(ctx, t, nodeEndpoint)

			err = aioContainer.Terminate(ctx)
			require.NoError(t, err)
			cancel()
			<-ctx.Done()
		})
	}
}

func runTests(_ context.Context, t *testing.T, nodeEndpoint string) {
	nodeAddr := "grpc://" + nodeEndpoint

	t.Run("PoolInterfaceWithAIO", func(t *testing.T) {
		testPoolInterfaceWithAIO(t, nodeAddr)
	})
}

func createDockerContainer(ctx context.Context, t *testing.T, image string) testcontainers.Container {
	const restGWListenEndpoint = "0.0.0.0:8090"
	req := testcontainers.ContainerRequest{
		Image: image,
		// timeout is chosen to have enough time for NeoFS chain deployment from scratch within NeoFS AIO
		WaitingFor:   wait.NewLogStrategy("aio container started").WithStartupTimeout(2 * time.Minute),
		Name:         "sdk-poll-tests-" + strconv.FormatInt(time.Now().UnixNano(), 36),
		Hostname:     "aio_autotest_" + strconv.FormatInt(time.Now().UnixNano(), 36),
		ExposedPorts: []string{"8080/tcp"},
		Env: map[string]string{
			"REST_GW_WALLET_PATH":                "/config/wallet-rest.json",
			"REST_GW_WALLET_PASSPHRASE":          "one",
			"REST_GW_WALLET_ADDRESS":             "NPFCqWHfi9ixCJRu7DABRbVfXRbkSEr9Vo",
			"REST_GW_POOL_PEERS_0_ADDRESS":       "localhost:8080",
			"REST_GW_SERVER_ENDPOINTS_0_ADDRESS": restGWListenEndpoint, // latest
		},
	}
	aioC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	tickNewEpoch = func(ctx context.Context, executor NetworkInfoExecutor) (int, error) {
		ni, err := executor.NetworkInfo(ctx, client.PrmNetworkInfo{})
		if err != nil {
			return 0, err
		}

		newEpoch := ni.CurrentEpoch() + 1

		_, _, err = aioC.Exec(ctx, tickEpochCmdBuilder(newEpoch))
		if err != nil {
			return 0, err
		}

		<-time.After(timeoutAfterEpochChange)
		return int(newEpoch), nil
	}

	return aioC
}

func testPoolInterfaceWithAIO(t *testing.T, nodeAddr string) {
	ctx := context.Background()

	account, signer, cont := testData(t)

	poolStat := stat.NewPoolStatistic()
	opts := pool.DefaultOptions()
	opts.SetStatisticCallback(poolStat.OperationCallback)
	opts.SetSessionExpirationDuration(sessionExpirationInEpochs)
	opts.SetClientRebalanceInterval(clientRebalanceInterval)

	pl, err := pool.New(pool.NewFlatNodeParams([]string{nodeAddr}), signer, opts)
	require.NoError(t, err)
	require.NoError(t, pl.Dial(ctx))

	t.Run("balance ok", func(t *testing.T) {
		var cmd client.PrmBalanceGet
		cmd.SetAccount(account)
		_, err = pl.BalanceGet(ctx, cmd)
		require.NoError(t, err)

		st := poolStat.Statistic()
		nodeStat, err := st.Node(nodeAddr)
		require.NoError(t, err)

		snap, err := nodeStat.Snapshot(stat.MethodBalanceGet)
		require.NoError(t, err)

		require.Equal(t, uint64(1), snap.AllRequests())
		require.Greater(t, snap.AllTime(), uint64(0))
	})

	t.Run("balance err", func(t *testing.T) {
		var id user.ID

		var cmd client.PrmBalanceGet
		cmd.SetAccount(id)
		_, err = pl.BalanceGet(ctx, cmd)
		require.Error(t, err)

		st := poolStat.Statistic()
		nodeStat, err := st.Node(nodeAddr)
		require.NoError(t, err)

		snap, err := nodeStat.Snapshot(stat.MethodBalanceGet)
		require.NoError(t, err)

		require.Equal(t, uint64(1), nodeStat.OverallErrors())
		require.Equal(t, uint64(2), snap.AllRequests())
		require.Greater(t, snap.AllTime(), uint64(0))
	})

	t.Run("container", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(context.Background(), defaultTimeOut)
		t.Cleanup(cancel)

		containerID := testCreateContainer(ctxTimeout, t, signer, cont, pl)
		cl, err := pl.RawClient()

		require.NoError(t, err)
		require.NoError(t, isContainerCreated(ctxTimeout, cl, containerID))

		t.Run("set eacl", func(t *testing.T) {
			ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
			t.Cleanup(cancel)

			containerID := testCreateContainer(ctxTimeout, t, signer, cont, pl)

			cnt, err := pl.ContainerGet(ctx, containerID, client.PrmContainerGet{})
			require.NoError(t, err)

			eaclTable := testSetEacl(ctxTimeout, t, signer, testEaclTable(containerID), pl, cnt.Revision())
			cl, err := pl.RawClient()

			require.NoError(t, err)
			require.NoError(t, isEACLCreated(ctxTimeout, cl, containerID, eaclTable))
		})
		t.Run("objects", func(t *testing.T) {
			type createObjFunc = func(ctx context.Context, t *testing.T, account user.ID, containerID cid.ID, signer user.Signer, payload []byte, putter objectPutIniter) oid.ID
			var creators = []createObjFunc{
				testObjectPutInit,
				testObjectPutInitReaderFrom,
			}

			for _, creator := range creators {
				payload := testutil.RandByteSlice(8)

				ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
				t.Cleanup(cancel)

				objectID := creator(ctxTimeout, t, account, containerID, signer, payload, pl)

				t.Run("download", func(t *testing.T) {
					ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
					t.Cleanup(cancel)

					var cmd client.PrmObjectGet

					hdr, read, err := pl.ObjectGetInit(ctxTimeout, containerID, objectID, signer, cmd)
					require.NoError(t, err)
					t.Cleanup(func() { _ = read.Close() })

					require.False(t, hdr.Owner().IsZero())
					require.True(t, hdr.Owner() == account)

					downloadedPayload := make([]byte, len(payload))

					l, err := read.Read(downloadedPayload)
					require.NoError(t, err)
					require.Equal(t, l, len(payload))

					require.True(t, bytes.Equal(payload, downloadedPayload))
				})
				t.Run("delete", func(t *testing.T) {
					ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
					t.Cleanup(cancel)

					testDeleteObject(ctxTimeout, t, signer, containerID, objectID, pl)
					cl, err := pl.RawClient()

					require.NoError(t, err)
					require.NoError(t, isObjectDeleted(ctxTimeout, cl, containerID, objectID, signer))
				})
				t.Run("epochs", func(t *testing.T) {
					var objectList []oid.ID
					times := int(sessionExpirationInEpochs * 3)
					for range times {
						epoch, err := tickNewEpoch(ctx, pl)
						require.NoError(t, err)

						t.Run(fmt.Sprintf("upload at epoch#%d", epoch), func(t *testing.T) {
							ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut*time.Duration(times))
							t.Cleanup(cancel)

							payload = append(payload, 0x01) // Make it different from the one above, otherwise OID will be the same and we can get "status: code = 2052 message = object already removed"
							objID := testObjectPutInit(ctxTimeout, t, account, containerID, signer, payload, pl)
							objectList = append(objectList, objID)
						})
					}
					for _, objID := range objectList {
						epoch, err := tickNewEpoch(ctx, pl)
						require.NoError(t, err)

						t.Run(fmt.Sprintf("delete at epoch#%d", epoch), func(t *testing.T) {
							ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut*time.Duration(times))
							t.Cleanup(cancel)

							testDeleteObject(ctxTimeout, t, signer, containerID, objID, pl)

							cl, err := pl.RawClient()
							require.NoError(t, err)

							require.NoError(t, isObjectDeleted(ctxTimeout, cl, containerID, objID, signer))
						})
					}
				})
			}
		})
		t.Run("delete", func(t *testing.T) {
			ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
			t.Cleanup(cancel)

			testDeleteContainer(ctxTimeout, t, signer, containerID, pl)
			cl, err := pl.RawClient()

			require.NoError(t, err)
			require.NoError(t, isContainerDeleted(ctxTimeout, cl, containerID))
		})
	})

	t.Run("search", func(t *testing.T) {
		testSearchObjectsWithAIO(ctx, t, pl, poolStat, nodeAddr, signer)
	})
}

func sortOIDs(ids []oid.ID) []oid.ID {
	sorted := slices.Clone(ids)
	slices.SortFunc(sorted, oid.ID.Compare)
	return sorted
}

func extractOIDs(items []client.SearchResultItem) []oid.ID {
	ids := make([]oid.ID, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	return ids
}

func testCreateContainerWithCleanup(ctx context.Context, t *testing.T, pl *pool.Pool, cl *client.Client, signer user.Signer, cont container.Container) cid.ID {
	ctxCreate, cancelCreate := context.WithTimeout(ctx, defaultTimeOut)
	defer cancelCreate()

	containerID := testCreateContainer(ctxCreate, t, signer, cont, pl)
	require.NoError(t, isContainerCreated(ctxCreate, cl, containerID))
	t.Cleanup(func() {
		ctxCleanup, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		testDeleteContainer(ctxCleanup, t, signer, containerID, pl)
		require.NoError(t, isContainerDeleted(ctxCleanup, cl, containerID))
	})

	return containerID
}

func testSearchStat(t *testing.T, poolStat *stat.PoolStat, nodeAddr string) (uint64, uint64, uint64) {
	st := poolStat.Statistic()
	nodeStat, err := st.Node(nodeAddr)
	require.NoError(t, err)

	snap, err := nodeStat.Snapshot(stat.MethodObjectSearchV2)
	require.NoError(t, err)

	return snap.AllRequests(), snap.AllTime(), nodeStat.OverallErrors()
}

func testSearchObjectsWithAIO(ctx context.Context, t *testing.T, pl *pool.Pool, poolStat *stat.PoolStat, nodeAddr string, signer user.Signer) {
	account := signer.UserID()

	cl, err := pl.RawClient()
	require.NoError(t, err)

	containerID := testCreateContainerWithCleanup(ctx, t, pl, cl, signer, testContainer(account, acl.PublicRWExtended))

	t.Run("empty", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		t.Cleanup(cancel)

		reqBefore, timeBefore, _ := testSearchStat(t, poolStat, nodeAddr)

		items, cursor, err := pl.SearchObjects(ctxTimeout, containerID, nil, nil, "", signer, client.SearchObjectsOptions{})
		require.NoError(t, err)
		require.Empty(t, items)
		require.Empty(t, cursor)

		reqAfter, timeAfter, _ := testSearchStat(t, poolStat, nodeAddr)
		require.Equal(t, reqBefore+1, reqAfter)
		require.Greater(t, timeAfter, timeBefore)
	})

	type fixtureSpec struct {
		size  int
		tag   string
		group string
		score string
		opt   string
	}
	specs := []fixtureSpec{
		{size: 10, tag: "alpha-1", group: "g1", score: "2", opt: "present-0"},
		{size: 20, tag: "alpha-2", group: "g1", score: "10"},
		{size: 30, tag: "beta-1", group: "g2", score: "100", opt: "present-2"},
		{size: 40, tag: "alpha-3", group: "g2", score: "20"},
		{size: 50, tag: "gamma", group: "g1", score: "3", opt: "present-4"},
	}

	ctxUpload, cancelUpload := context.WithTimeout(ctx, defaultTimeOut)
	defer cancelUpload()

	ids := make([]oid.ID, len(specs))
	for i, sp := range specs {
		payload := bytes.Repeat([]byte{byte(i + 1)}, sp.size)
		attrs := []object.Attribute{
			object.NewAttribute("tag", sp.tag),
			object.NewAttribute("group", sp.group),
			object.NewAttribute("score", sp.score),
		}
		if sp.opt != "" {
			attrs = append(attrs, object.NewAttribute("opt", sp.opt))
		}
		ids[i] = testObjectPutInitWithHeader(ctxUpload, t, account, containerID, signer, payload, pl, func(hdr *object.Object) {
			hdr.SetAttributes(attrs...)
		})
	}
	sortedAllOIDs := sortOIDs(ids)

	t.Run("pagination", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		t.Cleanup(cancel)

		var opts client.SearchObjectsOptions
		opts.SetCount(2)

		var (
			got    []oid.ID
			cursor string
			pages  int
		)
		for {
			items, nextCursor, err := pl.SearchObjects(ctxTimeout, containerID, nil, nil, cursor, signer, opts)
			require.NoError(t, err)
			require.LessOrEqual(t, len(items), 2)

			for _, item := range items {
				require.Empty(t, item.Attributes)
				got = append(got, item.ID)
			}
			pages++

			if nextCursor == "" {
				break
			}
			require.NotEqual(t, cursor, nextCursor)
			cursor = nextCursor
		}

		require.Equal(t, 3, pages)
		require.Equal(t, sortedAllOIDs, got)
	})

	t.Run("filters and attributes", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		t.Cleanup(cancel)

		t.Run("string equal", func(t *testing.T) {
			var fs object.SearchFilters
			fs.AddFilter("tag", "beta-1", object.MatchStringEqual)

			items, cursor, err := pl.SearchObjects(ctxTimeout, containerID, fs, nil, "", signer, client.SearchObjectsOptions{})
			require.NoError(t, err)
			require.Empty(t, cursor)
			require.Equal(t, []client.SearchResultItem{{ID: ids[2]}}, items)
		})

		t.Run("string not equal", func(t *testing.T) {
			var fs object.SearchFilters
			fs.AddFilter("tag", "gamma", object.MatchStringNotEqual)

			items, cursor, err := pl.SearchObjects(ctxTimeout, containerID, fs, []string{"tag"}, "", signer, client.SearchObjectsOptions{})
			require.NoError(t, err)
			require.Empty(t, cursor)
			require.Equal(t, []client.SearchResultItem{
				{ID: ids[0], Attributes: []string{"alpha-1"}},
				{ID: ids[1], Attributes: []string{"alpha-2"}},
				{ID: ids[3], Attributes: []string{"alpha-3"}},
				{ID: ids[2], Attributes: []string{"beta-1"}},
			}, items)
		})

		t.Run("common prefix with optional attribute", func(t *testing.T) {
			var fs object.SearchFilters
			fs.AddFilter("tag", "alpha-", object.MatchCommonPrefix)

			items, cursor, err := pl.SearchObjects(ctxTimeout, containerID, fs, []string{"tag", "opt"}, "", signer, client.SearchObjectsOptions{})
			require.NoError(t, err)
			require.Empty(t, cursor)
			require.Equal(t, []client.SearchResultItem{
				{ID: ids[0], Attributes: []string{"alpha-1", "present-0"}},
				{ID: ids[1], Attributes: []string{"alpha-2", ""}},
				{ID: ids[3], Attributes: []string{"alpha-3", ""}},
			}, items)
		})

		t.Run("not present", func(t *testing.T) {
			var fs object.SearchFilters
			fs.AddFilter("opt", "", object.MatchNotPresent)

			items, cursor, err := pl.SearchObjects(ctxTimeout, containerID, fs, nil, "", signer, client.SearchObjectsOptions{})
			require.NoError(t, err)
			require.Empty(t, cursor)
			require.Equal(t, sortOIDs([]oid.ID{ids[1], ids[3]}), extractOIDs(items))
		})

		t.Run("multi-filter conjunction", func(t *testing.T) {
			var fs object.SearchFilters
			fs.AddFilter("tag", "alpha-", object.MatchCommonPrefix)
			fs.AddFilter("group", "g1", object.MatchStringEqual)

			items, cursor, err := pl.SearchObjects(ctxTimeout, containerID, fs, []string{"tag", "group"}, "", signer, client.SearchObjectsOptions{})
			require.NoError(t, err)
			require.Empty(t, cursor)
			require.Equal(t, []client.SearchResultItem{
				{ID: ids[0], Attributes: []string{"alpha-1", "g1"}},
				{ID: ids[1], Attributes: []string{"alpha-2", "g1"}},
			}, items)
		})

		t.Run("numeric matchers", func(t *testing.T) {
			cases := []struct {
				name string
				op   object.SearchMatchType
				val  string
				want []client.SearchResultItem
			}{
				{
					name: "gt",
					op:   object.MatchNumGT,
					val:  "2",
					want: []client.SearchResultItem{
						{ID: ids[4], Attributes: []string{"3"}},
						{ID: ids[1], Attributes: []string{"10"}},
						{ID: ids[3], Attributes: []string{"20"}},
						{ID: ids[2], Attributes: []string{"100"}},
					},
				},
				{
					name: "ge",
					op:   object.MatchNumGE,
					val:  "10",
					want: []client.SearchResultItem{
						{ID: ids[1], Attributes: []string{"10"}},
						{ID: ids[3], Attributes: []string{"20"}},
						{ID: ids[2], Attributes: []string{"100"}},
					},
				},
				{
					name: "lt",
					op:   object.MatchNumLT,
					val:  "20",
					want: []client.SearchResultItem{
						{ID: ids[0], Attributes: []string{"2"}},
						{ID: ids[4], Attributes: []string{"3"}},
						{ID: ids[1], Attributes: []string{"10"}},
					},
				},
				{
					name: "le",
					op:   object.MatchNumLE,
					val:  "10",
					want: []client.SearchResultItem{
						{ID: ids[0], Attributes: []string{"2"}},
						{ID: ids[4], Attributes: []string{"3"}},
						{ID: ids[1], Attributes: []string{"10"}},
					},
				},
			}

			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					var fs object.SearchFilters
					fs.AddFilter("score", tc.val, tc.op)

					items, cursor, err := pl.SearchObjects(ctxTimeout, containerID, fs, []string{"score"}, "", signer, client.SearchObjectsOptions{})
					require.NoError(t, err)
					require.Empty(t, cursor)
					require.Equal(t, tc.want, items)
				})
			}

			t.Run("range", func(t *testing.T) {
				var fs object.SearchFilters
				fs.AddFilter("score", "3", object.MatchNumGE)
				fs.AddFilter("score", "20", object.MatchNumLE)

				items, cursor, err := pl.SearchObjects(ctxTimeout, containerID, fs, []string{"score"}, "", signer, client.SearchObjectsOptions{})
				require.NoError(t, err)
				require.Empty(t, cursor)
				require.Equal(t, []client.SearchResultItem{
					{ID: ids[4], Attributes: []string{"3"}},
					{ID: ids[1], Attributes: []string{"10"}},
					{ID: ids[3], Attributes: []string{"20"}},
				}, items)
			})
		})
	})

	t.Run("system filters", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		t.Cleanup(cancel)

		t.Run("payload size", func(t *testing.T) {
			var fs object.SearchFilters
			fs.AddPayloadSizeFilter(object.MatchNumGE, 30)

			items, cursor, err := pl.SearchObjects(ctxTimeout, containerID, fs, []string{object.FilterPayloadSize}, "", signer, client.SearchObjectsOptions{})
			require.NoError(t, err)
			require.Empty(t, cursor)
			require.Equal(t, []client.SearchResultItem{
				{ID: ids[2], Attributes: []string{"30"}},
				{ID: ids[3], Attributes: []string{"40"}},
				{ID: ids[4], Attributes: []string{"50"}},
			}, items)
		})

		t.Run("creation epoch", func(t *testing.T) {
			headFirst, err := pl.ObjectHead(ctxTimeout, containerID, ids[0], signer, client.PrmObjectHead{})
			require.NoError(t, err)
			headLast, err := pl.ObjectHead(ctxTimeout, containerID, ids[len(ids)-1], signer, client.PrmObjectHead{})
			require.NoError(t, err)

			var fs object.SearchFilters
			fs.AddCreationEpochFilter(object.MatchNumGE, headFirst.CreationEpoch())
			fs.AddCreationEpochFilter(object.MatchNumLE, headLast.CreationEpoch())

			items, cursor, err := pl.SearchObjects(ctxTimeout, containerID, fs, nil, "", signer, client.SearchObjectsOptions{})
			require.NoError(t, err)
			require.Empty(t, cursor)
			require.ElementsMatch(t, ids, extractOIDs(items))

			var fsEmpty object.SearchFilters
			fsEmpty.AddCreationEpochFilter(object.MatchNumGT, headLast.CreationEpoch())

			items, cursor, err = pl.SearchObjects(ctxTimeout, containerID, fsEmpty, nil, "", signer, client.SearchObjectsOptions{})
			require.NoError(t, err)
			require.Empty(t, cursor)
			require.Empty(t, items)
		})

		t.Run("owner ID", func(t *testing.T) {
			var fs object.SearchFilters
			fs.AddObjectOwnerIDFilter(object.MatchStringEqual, account)

			items, cursor, err := pl.SearchObjects(ctxTimeout, containerID, fs, nil, "", signer, client.SearchObjectsOptions{})
			require.NoError(t, err)
			require.Empty(t, cursor)
			require.Equal(t, sortedAllOIDs, extractOIDs(items))

			var fsOther object.SearchFilters
			fsOther.AddObjectOwnerIDFilter(object.MatchStringEqual, usertest.User().RFC6979.UserID())

			items, cursor, err = pl.SearchObjects(ctxTimeout, containerID, fsOther, nil, "", signer, client.SearchObjectsOptions{})
			require.NoError(t, err)
			require.Empty(t, cursor)
			require.Empty(t, items)
		})

		t.Run("type", func(t *testing.T) {
			var fs object.SearchFilters
			fs.AddTypeFilter(object.MatchStringEqual, object.TypeRegular)

			items, cursor, err := pl.SearchObjects(ctxTimeout, containerID, fs, nil, "", signer, client.SearchObjectsOptions{})
			require.NoError(t, err)
			require.Empty(t, cursor)
			require.Equal(t, sortedAllOIDs, extractOIDs(items))

			var fsLock object.SearchFilters
			fsLock.AddTypeFilter(object.MatchStringEqual, object.TypeLock)

			items, cursor, err = pl.SearchObjects(ctxTimeout, containerID, fsLock, nil, "", signer, client.SearchObjectsOptions{})
			require.NoError(t, err)
			require.Empty(t, cursor)
			require.Empty(t, items)
		})

		t.Run("root", func(t *testing.T) {
			var fs object.SearchFilters
			fs.AddRootFilter()

			items, cursor, err := pl.SearchObjects(ctxTimeout, containerID, fs, nil, "", signer, client.SearchObjectsOptions{})
			require.NoError(t, err)
			require.Empty(t, cursor)
			require.Equal(t, sortedAllOIDs, extractOIDs(items))
		})

		t.Run("phy", func(t *testing.T) {
			var fs object.SearchFilters
			fs.AddPhyFilter()

			items, cursor, err := pl.SearchObjects(ctxTimeout, containerID, fs, nil, "", signer, client.SearchObjectsOptions{})
			require.NoError(t, err)
			require.Empty(t, cursor)
			require.Equal(t, sortedAllOIDs, extractOIDs(items))
		})
	})

	t.Run("local", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		t.Cleanup(cancel)

		var opts client.SearchObjectsOptions
		opts.DisableForwarding()

		t.Run("equality filter", func(t *testing.T) {
			var fs object.SearchFilters
			fs.AddFilter("tag", "beta-1", object.MatchStringEqual)

			items, cursor, err := cl.SearchObjects(ctxTimeout, containerID, fs, nil, "", signer, opts)
			require.NoError(t, err)
			require.Empty(t, cursor)
			require.Equal(t, []client.SearchResultItem{{ID: ids[2]}}, items)
		})

		t.Run("non-equality filter without requested attributes", func(t *testing.T) {
			var fs object.SearchFilters
			fs.AddFilter("tag", "alpha-", object.MatchCommonPrefix)

			items, cursor, err := cl.SearchObjects(ctxTimeout, containerID, fs, nil, "", signer, opts)
			require.NoError(t, err)
			require.Empty(t, cursor)
			require.Equal(t, []client.SearchResultItem{
				{ID: ids[0], Attributes: []string{"alpha-1"}},
				{ID: ids[1], Attributes: []string{"alpha-2"}},
				{ID: ids[3], Attributes: []string{"alpha-3"}},
			}, items)
		})
	})

	t.Run("container revision", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		t.Cleanup(cancel)

		cnt, err := pl.ContainerGet(ctxTimeout, containerID, client.PrmContainerGet{})
		require.NoError(t, err)

		if rev := cnt.Revision(); rev > 0 {
			var okOpts client.SearchObjectsOptions
			okOpts.AttachContainerRevision(rev)

			items, cursor, err := pl.SearchObjects(ctxTimeout, containerID, nil, nil, "", signer, okOpts)
			require.NoError(t, err)
			require.Empty(t, cursor)
			require.Equal(t, sortedAllOIDs, extractOIDs(items))

			var badOpts client.SearchObjectsOptions
			badOpts.AttachContainerRevision(rev + 1)

			_, _, err = pl.SearchObjects(ctxTimeout, containerID, nil, nil, "", signer, badOpts)
			require.ErrorIs(t, err, apistatus.ErrContainerRevisionMismatch)
		} else {
			var revOpts client.SearchObjectsOptions
			revOpts.AttachContainerRevision(1)

			_, _, err = pl.SearchObjects(ctxTimeout, containerID, nil, nil, "", signer, revOpts)
			require.ErrorContains(t, err, "container revision was attached, but server does not support it")
		}
	})

	t.Run("access control", func(t *testing.T) {
		testSearchObjectsAccessControlWithAIO(ctx, t, pl, cl, poolStat, nodeAddr, signer)
	})

	t.Run("deleted object", func(t *testing.T) {
		ctxDel, cancelDel := context.WithTimeout(ctx, defaultTimeOut)
		t.Cleanup(cancelDel)

		deletedID := ids[len(ids)-1]
		expectedRemaining := sortOIDs(ids[:len(ids)-1])

		testDeleteObject(ctxDel, t, signer, containerID, deletedID, pl)
		require.NoError(t, isObjectDeleted(ctxDel, cl, containerID, deletedID, signer))

		ctxSearch, cancelSearch := context.WithTimeout(ctx, defaultTimeOut)
		t.Cleanup(cancelSearch)

		var rootFilter object.SearchFilters
		rootFilter.AddRootFilter()

		require.Eventually(t, func() bool {
			allItems, cursor, err := pl.SearchObjects(ctxSearch, containerID, nil, nil, "", signer, client.SearchObjectsOptions{})
			if err != nil || cursor != "" || slices.Contains(extractOIDs(allItems), deletedID) {
				return false
			}

			rootItems, cursor, err := pl.SearchObjects(ctxSearch, containerID, rootFilter, nil, "", signer, client.SearchObjectsOptions{})
			if err != nil || cursor != "" {
				return false
			}
			return slices.Equal(expectedRemaining, extractOIDs(rootItems))
		}, defaultTimeOut, tickInterval)
	})

	t.Run("expiration time", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		t.Cleanup(cancel)

		epInfo, err := cl.EndpointInfo(ctxTimeout, client.PrmEndpointInfo{})
		require.NoError(t, err)
		if epInfo.LatestVersion().Compare(version.New(2, 28)) < 0 {
			t.Skipf("node API %s does not support expiration time filter", epInfo.LatestVersion())
		}

		expTime := time.Now().Add(2 * time.Hour).Truncate(time.Second)
		expObjID := testObjectPutInitWithHeader(ctxTimeout, t, account, containerID, signer, []byte("expiring-obj"), pl, func(hdr *object.Object) {
			require.NoError(t, hdr.SetExpirationTime(expTime))
		})

		var fs object.SearchFilters
		fs.AddExpirationTimeFilter(object.MatchNumGE, expTime)

		items, cursor, err := pl.SearchObjects(ctxTimeout, containerID, fs, []string{object.FilterExpirationTime}, "", signer, client.SearchObjectsOptions{})
		require.NoError(t, err)
		require.Empty(t, cursor)
		require.Equal(t, []client.SearchResultItem{
			{ID: expObjID, Attributes: []string{strconv.FormatInt(expTime.Unix(), 10)}},
		}, items)

		var fsAfter object.SearchFilters
		fsAfter.AddExpirationTimeFilter(object.MatchNumGT, expTime)

		items, cursor, err = pl.SearchObjects(ctxTimeout, containerID, fsAfter, nil, "", signer, client.SearchObjectsOptions{})
		require.NoError(t, err)
		require.Empty(t, cursor)
		require.Empty(t, items)
	})
}

func testSearchObjectsAccessControlWithAIO(ctx context.Context, t *testing.T, pl *pool.Pool, cl *client.Client, poolStat *stat.PoolStat, nodeAddr string, signer user.Signer) {
	account := signer.UserID()
	aclCnrID := testCreateContainerWithCleanup(ctx, t, pl, cl, signer, testContainer(account, acl.PublicRWExtended))

	ctxEACL, cancelEACL := context.WithTimeout(ctx, defaultTimeOut)
	defer cancelEACL()

	aclCnt, err := pl.ContainerGet(ctxEACL, aclCnrID, client.PrmContainerGet{})
	require.NoError(t, err)

	denyTable := testEaclOthersTable(aclCnrID, eacl.OperationSearch, eacl.ActionDeny)
	testSetEacl(ctxEACL, t, signer, denyTable, pl, aclCnt.Revision())
	require.NoError(t, isEACLCreated(ctxEACL, cl, aclCnrID, denyTable))

	aclObjID := testObjectPutInit(ctxEACL, t, account, aclCnrID, signer, []byte("acl-obj"), pl)

	otherSigner := usertest.User().RFC6979

	ni, err := pl.NetworkInfo(ctxEACL, client.PrmNetworkInfo{})
	require.NoError(t, err)
	curEpoch := ni.CurrentEpoch()

	t.Run("unauthorized", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		t.Cleanup(cancel)

		reqBefore, _, errBefore := testSearchStat(t, poolStat, nodeAddr)

		_, _, err := pl.SearchObjects(ctxTimeout, aclCnrID, nil, nil, "", otherSigner, client.SearchObjectsOptions{})
		require.ErrorIs(t, err, apistatus.ErrObjectAccessDenied)

		reqAfter, _, errAfter := testSearchStat(t, poolStat, nodeAddr)
		require.Equal(t, reqBefore+1, reqAfter)
		require.Equal(t, errBefore+1, errAfter)
	})

	t.Run("bearer token", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		t.Cleanup(cancel)

		var bt bearer.Token
		bt.SetEACLTable(testEaclOthersTable(aclCnrID, eacl.OperationSearch, eacl.ActionAllow))
		bt.ForUser(otherSigner.UserID())
		bt.SetIat(curEpoch)
		bt.SetNbf(curEpoch)
		bt.SetExp(curEpoch + 10)
		require.NoError(t, bt.Sign(signer))

		var opts client.SearchObjectsOptions
		opts.WithBearerToken(bt)

		items, cursor, err := pl.SearchObjects(ctxTimeout, aclCnrID, nil, nil, "", otherSigner, opts)
		require.NoError(t, err)
		require.Empty(t, cursor)
		require.Equal(t, []client.SearchResultItem{{ID: aclObjID}}, items)
	})

	t.Run("session token v2", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		t.Cleanup(cancel)

		epInfo, err := cl.EndpointInfo(ctxTimeout, client.PrmEndpointInfo{})
		require.NoError(t, err)
		if epInfo.LatestVersion().Compare(version.New(2, 22)) < 0 {
			t.Skipf("node API %s does not support session token v2 in request meta header", epInfo.LatestVersion())
		}

		var stV2 sessionv2.Token
		stV2.SetVersion(0) // nodes prior to API 2.29 only accept token version 0

		now := time.Now()
		stV2.SetIat(now.Add(-time.Minute))
		stV2.SetNbf(now.Add(-time.Minute))
		stV2.SetExp(now.Add(time.Duration(sessionExpirationInSeconds) * time.Second))
		require.NoError(t, stV2.AddSubject(sessionv2.NewTargetUser(otherSigner.UserID())))

		sCtx, err := sessionv2.NewContext(aclCnrID, []sessionv2.Verb{sessionv2.VerbObjectSearch})
		require.NoError(t, err)
		require.NoError(t, stV2.AddContext(sCtx))
		require.NoError(t, stV2.Sign(signer))

		var opts client.SearchObjectsOptions
		opts.WithSessionTokenV2(stV2)

		items, cursor, err := pl.SearchObjects(ctxTimeout, aclCnrID, nil, nil, "", otherSigner, opts)
		require.NoError(t, err)
		require.Empty(t, cursor)
		require.Equal(t, []client.SearchResultItem{{ID: aclObjID}}, items)
	})
}

func testObjectPutInit(ctx context.Context, t *testing.T, account user.ID, containerID cid.ID, signer user.Signer, payload []byte, putter objectPutIniter) oid.ID {
	return testObjectPutInitWithHeader(ctx, t, account, containerID, signer, payload, putter, nil)
}

func testObjectPutInitWithHeader(ctx context.Context, t *testing.T, account user.ID, containerID cid.ID, signer user.Signer, payload []byte, putter objectPutIniter, modify func(*object.Object)) oid.ID {
	hdr := object.New(containerID, account)
	if modify != nil {
		modify(hdr)
	}

	var prm client.PrmObjectPutInit

	w, err := putter.ObjectPutInit(ctx, *hdr, signer, prm)
	require.NoError(t, err)

	_, err = w.Write(payload)
	require.NoError(t, err)

	err = w.Close()
	require.NoError(t, err)

	return w.GetResult().StoredObjectID()
}

func testObjectPutInitReaderFrom(ctx context.Context, t *testing.T, account user.ID, containerID cid.ID, signer user.Signer, payload []byte, putter objectPutIniter) oid.ID {
	var hdr = object.New(containerID, account)

	var prm client.PrmObjectPutInit

	w, err := putter.ObjectPutInit(ctx, *hdr, signer, prm)
	require.NoError(t, err)

	n, err := w.ReadFrom(bytes.NewReader(payload))
	require.Equal(t, n, int64(len(payload)))
	require.NoError(t, err)

	err = w.Close()
	require.NoError(t, err)

	return w.GetResult().StoredObjectID()
}

func testCreateContainer(ctx context.Context, t *testing.T, signer neofscrypto.Signer, cont container.Container, creator containerCreator) cid.ID {
	var rd netmap.ReplicaDescriptor
	rd.SetNumberOfObjects(1)

	var pp netmap.PlacementPolicy
	pp.SetContainerBackupFactor(1)
	pp.SetReplicas([]netmap.ReplicaDescriptor{rd})

	cont.SetPlacementPolicy(pp)

	var cmd client.PrmContainerPut

	containerID, err := creator.ContainerPut(ctx, cont, signer, cmd)
	require.NoError(t, err)

	return containerID
}

func testDeleteContainer(ctx context.Context, t *testing.T, signer neofscrypto.Signer, containerID cid.ID, deleter containerDeleter) {
	var cmd client.PrmContainerDelete

	require.NoError(t, deleter.ContainerDelete(ctx, containerID, signer, cmd))
}

func testDeleteObject(ctx context.Context, t *testing.T, signer user.Signer, containerID cid.ID, objectID oid.ID, deleter objectDeleter) {
	var cmd client.PrmObjectDelete

	_, err := deleter.ObjectDelete(ctx, containerID, objectID, signer, cmd)
	require.NoError(t, err)
}

func testSetEacl(ctx context.Context, t *testing.T, signer user.Signer, table eacl.Table, setter containerEaclSetter, revision uint64) eacl.Table {
	var prm client.PrmContainerSetEACL

	if revision > 0 {
		prm.AttachContainerRevision(revision)
	}

	require.NoError(t, setter.ContainerSetEACL(ctx, table, signer, prm))

	return table
}

func testGetEacl(ctx context.Context, t *testing.T, containerID cid.ID, table eacl.Table, setter containerEaclGetter) {
	var prm client.PrmContainerEACL

	newTable, err := setter.ContainerEACL(ctx, containerID, prm)
	require.NoError(t, err)
	require.Equal(t, table.Marshal(), newTable.Marshal())
}

func isContainerCreated(ctx context.Context, c containerGetter, id cid.ID) error {
	t := time.NewTicker(tickInterval)

	var cmdGet client.PrmContainerGet

	for {
		select {
		case <-t.C:
			_, err := c.ContainerGet(ctx, id, cmdGet)
			if err != nil {
				if errors.Is(err, apistatus.ErrContainerNotFound) {
					continue
				}

				return fmt.Errorf("ContainerGet %w", err)
			}
			return nil

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func isContainerDeleted(ctx context.Context, c containerGetter, id cid.ID) error {
	t := time.NewTicker(tickInterval)

	var cmdGet client.PrmContainerGet

	for {
		select {
		case <-t.C:
			_, err := c.ContainerGet(ctx, id, cmdGet)
			if err != nil {
				if errors.Is(err, apistatus.ErrContainerNotFound) {
					return nil
				}

				return fmt.Errorf("ContainerGet %w", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func isEACLCreated(ctx context.Context, c containerEaclGetter, id cid.ID, oldTable eacl.Table) error {
	oldBinary := oldTable.Marshal()

	t := time.NewTicker(tickInterval)

	var cmdGet client.PrmContainerEACL

	for {
		select {
		case <-t.C:
			table, err := c.ContainerEACL(ctx, id, cmdGet)
			if err != nil {
				if errors.Is(err, apistatus.ErrEACLNotFound) {
					continue
				}

				return fmt.Errorf("ContainerEACL %w", err)
			}

			if bytes.Equal(oldBinary, table.Marshal()) {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func isObjectDeleted(ctx context.Context, c objectHeadGetter, id cid.ID, oid oid.ID, signer user.Signer) error {
	t := time.NewTicker(tickInterval)

	var prmHead client.PrmObjectHead

	for {
		select {
		case <-t.C:
			_, err := c.ObjectHead(ctx, id, oid, signer, prmHead)
			if err != nil {
				if errors.Is(err, apistatus.ErrObjectNotFound) ||
					errors.Is(err, apistatus.ErrObjectAlreadyRemoved) {
					return nil
				}

				return fmt.Errorf("ObjectGetInit %w", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func createTokenV2(t *testing.T, ctx context.Context, cl *client.Client, containerID cid.ID, verb sessionv2.Verb, signer user.Signer) sessionv2.Token {
	var sess sessionv2.Token
	sess.SetVersion(sessionv2.TokenCurrentVersion)

	currentTime := time.Now()
	sess.SetIat(currentTime)
	sess.SetNbf(currentTime)
	sess.SetExp(currentTime.Add(time.Duration(sessionExpirationInSeconds) * time.Second))

	nm, err := cl.NetMapSnapshot(ctx, client.PrmNetMapSnapshot{})
	require.NoError(t, err)

	for _, node := range nm.Nodes() {
		neoPubKey, err := keys.NewPublicKeyFromBytes(node.PublicKey(), elliptic.P256())
		require.NoError(t, err)

		ecdsaPubKey := (*ecdsa.PublicKey)(neoPubKey)

		userID := user.NewFromECDSAPublicKey(*ecdsaPubKey)
		require.NoError(t, sess.AddSubject(sessionv2.NewTargetUser(userID)))
	}

	ctx2, err := sessionv2.NewContext(containerID, []sessionv2.Verb{verb})
	require.NoError(t, err)

	err = sess.AddContext(ctx2)
	require.NoError(t, err)

	err = sess.Sign(signer)
	require.NoError(t, err)
	return sess
}
