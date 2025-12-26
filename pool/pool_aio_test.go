//go:build aiotest

package pool_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/internal/testutil"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/pool"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/nspcc-dev/neofs-sdk-go/waiter"
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
		{image: "nspccdev/neofs-aio", version: "0.39.0"},
		{image: "nspccdev/neofs-aio", version: "latest"},
	}

	sessionExpirationInEpochs = uint64(2)

	// clientRebalanceInterval must be lower than timeoutAfterEpochChange.
	// It is important if you are forcing epoch changing. Otherwise, session tokens maybe expired and rejected by node.
	// It happens because session cache inside pool just wasn't updated yet.
	clientRebalanceInterval = 1 * time.Second
	timeoutAfterEpochChange = 2 * time.Second
)

type (
	newEpochTickerFunc func(context.Context, client.NetworkInfoExecutor) (int, error)

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

func testData(_ *testing.T) (user.ID, user.Signer, container.Container) {
	signer := usertest.User().RFC6979
	account := signer.UserID()

	containerName := strconv.FormatInt(time.Now().UnixNano(), 16)
	creationTime := time.Now()

	var cont container.Container
	cont.Init()
	cont.SetBasicACL(acl.PublicRWExtended)
	cont.SetOwner(account)
	cont.SetName(containerName)
	cont.SetCreationTime(creationTime)

	return account, signer, cont
}

func testEaclTable(containerID cid.ID) eacl.Table {
	var table eacl.Table
	table.SetCID(containerID)

	r := eacl.Record{}
	r.SetOperation(eacl.OperationPut)
	r.SetAction(eacl.ActionAllow)

	var target eacl.Target
	target.SetRole(eacl.RoleOthers)
	r.SetTargets(target)
	table.SetRecords([]eacl.Record{r})

	return table
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

	t.Run("PoolWaiterWithAIO", func(t *testing.T) {
		testPoolWaiterWithAIO(t, nodeAddr)
	})

	t.Run("ClientWaiterWithAIO", func(t *testing.T) {
		testClientWaiterWithAIO(t, nodeAddr)
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
			"REST_GW_LISTEN_ADDRESS":             restGWListenEndpoint, // 0.39.0
			"REST_GW_SERVER_ENDPOINTS_0_ADDRESS": restGWListenEndpoint, // latest
		},
	}
	aioC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Have to wait this time. Required for new tick event processing.
	// Should be removed after fix epochs in AIO start.
	<-time.After(3 * time.Second)

	_, _, err = aioC.Exec(ctx, tickEpochCmdBuilder(3))
	require.NoError(t, err)

	<-time.After(3 * time.Second)

	tickNewEpoch = func(ctx context.Context, executor client.NetworkInfoExecutor) (int, error) {
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

			eaclTable := testSetEacl(ctxTimeout, t, signer, testEaclTable(containerID), pl)
			cl, err := pl.RawClient()

			require.NoError(t, err)
			require.NoError(t, isEACLCreated(ctxTimeout, cl, containerID, eaclTable))
		})
		t.Run("objects", func(t *testing.T) {
			payload := testutil.RandByteSlice(8)

			ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
			t.Cleanup(cancel)

			objectID := testObjectPutInit(ctxTimeout, t, account, containerID, signer, payload, pl)

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
}

func testPoolWaiterWithAIO(t *testing.T, nodeAddr string) {
	ctx := context.Background()

	account, signer, cont := testData(t)

	pl, err := pool.New(pool.NewFlatNodeParams([]string{nodeAddr}), signer, pool.DefaultOptions())
	require.NoError(t, err)
	require.NoError(t, pl.Dial(ctx))

	wait := waiter.NewWaiter(pl, time.Second)

	t.Run("container", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		t.Cleanup(cancel)

		containerID := testCreateContainer(ctxTimeout, t, signer, cont, wait)

		t.Run("eacl", func(t *testing.T) {
			ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
			t.Cleanup(cancel)

			eaclTable := testSetEacl(ctxTimeout, t, signer, testEaclTable(containerID), wait)

			t.Run("get", func(t *testing.T) {
				ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
				t.Cleanup(cancel)

				testGetEacl(ctxTimeout, t, containerID, eaclTable, pl)
			})
		})
		t.Run("object", func(t *testing.T) {
			ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
			t.Cleanup(cancel)

			var hdr = object.New(containerID, account)

			var prm client.PrmObjectPutInit
			prm.SetCopiesNumber(1)

			w, err := pl.ObjectPutInit(ctxTimeout, *hdr, signer, prm)
			require.NoError(t, err)

			payload := testutil.RandByteSlice(8)
			_, err = w.Write(payload)
			require.NoError(t, err)

			err = w.Close()
			require.NoError(t, err)

			objectID := w.GetResult().StoredObjectID()

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
			})
		})
		t.Run("delete", func(t *testing.T) {
			ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
			t.Cleanup(cancel)

			testDeleteContainer(ctxTimeout, t, signer, containerID, wait)
		})
	})
}

func testClientWaiterWithAIO(t *testing.T, nodeAddr string) {
	ctx := context.Background()

	account, signer, cont := testData(t)

	var prmInit client.PrmInit

	cl, err := client.New(prmInit)
	require.NoError(t, err)

	// connect to NeoFS gateway
	var prmDial client.PrmDial
	prmDial.SetServerURI(nodeAddr) // endpoint address
	prmDial.SetTimeout(15 * time.Second)
	prmDial.SetStreamTimeout(15 * time.Second)

	err = cl.Dial(prmDial)
	require.NoError(t, err)

	wait := waiter.NewWaiter(cl, time.Second)

	t.Run("create container", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		t.Cleanup(cancel)

		containerID := testCreateContainer(ctxTimeout, t, signer, cont, wait)

		t.Run("eacl", func(t *testing.T) {
			ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
			t.Cleanup(cancel)

			eaclTable := testSetEacl(ctxTimeout, t, signer, testEaclTable(containerID), wait)

			t.Run("get eacl", func(t *testing.T) {
				ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
				t.Cleanup(cancel)

				testGetEacl(ctxTimeout, t, containerID, eaclTable, cl)
			})
		})
		t.Run("object", func(t *testing.T) {
			var prmSess client.PrmSessionCreate
			prmSess.SetExp(math.MaxUint64)

			res, err := cl.SessionCreate(ctx, signer, prmSess)
			require.NoError(t, err)

			var id uuid.UUID
			err = id.UnmarshalBinary(res.ID())
			require.NoError(t, err)

			var key neofsecdsa.PublicKey
			err = key.Decode(res.PublicKey())
			require.NoError(t, err)

			var sess session.Object

			sess.SetID(id)
			sess.SetAuthKey(&key)
			sess.SetExp(math.MaxUint64)
			sess.ForVerb(session.VerbObjectPut)
			sess.BindContainer(containerID)

			err = sess.Sign(signer)
			require.NoError(t, err)

			ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
			t.Cleanup(cancel)

			var hdr = object.New(containerID, account)

			var prm client.PrmObjectPutInit
			prm.SetCopiesNumber(1)
			prm.WithinSession(sess)

			w, err := cl.ObjectPutInit(ctxTimeout, *hdr, signer, prm)
			require.NoError(t, err)

			payload := testutil.RandByteSlice(8)
			_, err = w.Write(payload)
			require.NoError(t, err)

			err = w.Close()
			require.NoError(t, err)

			objectID := w.GetResult().StoredObjectID()

			t.Run("download", func(t *testing.T) {
				ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
				t.Cleanup(cancel)

				var cmd client.PrmObjectGet

				hdr, read, err := cl.ObjectGetInit(ctxTimeout, containerID, objectID, signer, cmd)
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

				testDeleteObject(ctxTimeout, t, signer, containerID, objectID, cl)
			})
		})
		t.Run("delete", func(t *testing.T) {
			ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
			t.Cleanup(cancel)

			testDeleteContainer(ctxTimeout, t, signer, containerID, wait)
		})
	})
}

func testObjectPutInit(ctx context.Context, t *testing.T, account user.ID, containerID cid.ID, signer user.Signer, payload []byte, putter objectPutIniter) oid.ID {
	var hdr = object.New(containerID, account)

	var prm client.PrmObjectPutInit
	prm.SetCopiesNumber(1)

	w, err := putter.ObjectPutInit(ctx, *hdr, signer, prm)
	require.NoError(t, err)

	_, err = w.Write(payload)
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

func testSetEacl(ctx context.Context, t *testing.T, signer user.Signer, table eacl.Table, setter containerEaclSetter) eacl.Table {
	var prm client.PrmContainerSetEACL

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
