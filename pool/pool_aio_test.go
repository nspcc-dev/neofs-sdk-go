//go:build aiotest

package pool

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
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
		{image: "nspccdev/neofs-aio", version: "0.36.0"},
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

func nodeAddress(nodeEndpoint string) string {
	return "grpc://" + nodeEndpoint
}

func getSigner() user.Signer {
	key, err := keys.NEP2Decrypt("6PYM8VdX2BSm7BSXKzV4Fz6S3R9cDLLWNrD9nMjxW352jEv3fsC8N3wNLY", "one", keys.NEP2ScryptParams())
	if err != nil {
		panic(err)
	}

	return user.NewAutoIDSignerRFC6979(key.PrivateKey)
}

func testData(_ *testing.T) (user.ID, user.Signer, container.Container) {
	signer := getSigner()
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

	r := eacl.NewRecord()
	r.SetOperation(eacl.OperationPut)
	r.SetAction(eacl.ActionAllow)

	var target eacl.Target
	target.SetRole(eacl.RoleOthers)
	r.SetTargets(target)
	table.AddRecord(r)

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
	nodeAddr := nodeAddress(nodeEndpoint)

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
	req := testcontainers.ContainerRequest{
		Image: image,
		// timeout is chosen to have enough time for NeoFS chain deployment from scratch within NeoFS AIO
		WaitingFor:   wait.NewLogStrategy("Serving neofs rest gw").WithStartupTimeout(2 * time.Minute),
		Name:         "sdk-poll-tests-" + strconv.FormatInt(time.Now().UnixNano(), 36),
		Hostname:     "aio_autotest_" + strconv.FormatInt(time.Now().UnixNano(), 36),
		ExposedPorts: []string{"8080/tcp"},
		Env: map[string]string{
			"REST_GW_WALLET_PATH":       "/config/wallet-rest.json",
			"REST_GW_WALLET_PASSPHRASE": "one",
			"REST_GW_WALLET_ADDRESS":    "NPFCqWHfi9ixCJRu7DABRbVfXRbkSEr9Vo",
			"REST_GW_PEERS_0_ADDRESS":   "localhost:8080",
			"REST_GW_LISTEN_ADDRESS":    "0.0.0.0:8090",
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
	var eaclTable eacl.Table

	poolStat := stat.NewPoolStatistic()
	opts := DefaultOptions()
	opts.SetStatisticCallback(poolStat.OperationCallback)
	opts.sessionExpirationDuration = sessionExpirationInEpochs
	opts.clientRebalanceInterval = clientRebalanceInterval

	pool, err := New(NewFlatNodeParams([]string{nodeAddr}), signer, opts)
	require.NoError(t, err)
	require.NoError(t, pool.Dial(ctx))

	var containerID cid.ID
	var objectID oid.ID
	var objectList []oid.ID

	payload := make([]byte, 8)
	_, err = rand.Read(payload)
	require.NoError(t, err)

	t.Run("balance ok", func(t *testing.T) {
		var cmd client.PrmBalanceGet
		cmd.SetAccount(account)
		_, err = pool.BalanceGet(ctx, cmd)
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
		_, err = pool.BalanceGet(ctx, cmd)
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

	t.Run("create container", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(context.Background(), defaultTimeOut)
		defer cancel()

		containerID = testCreateContainer(ctxTimeout, t, signer, cont, pool)
		cl, err := pool.sdkClient()

		require.NoError(t, err)
		require.NoError(t, isBucketCreated(ctxTimeout, cl, containerID))

		eaclTable = testEaclTable(containerID)
	})

	t.Run("set eacl", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		table := testSetEacl(ctxTimeout, t, signer, eaclTable, pool)
		cl, err := pool.sdkClient()

		require.NoError(t, err)
		require.NoError(t, isEACLCreated(ctxTimeout, cl, containerID, table))
	})

	t.Run("upload object", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		objectID = testObjectPutInit(ctxTimeout, t, account, containerID, signer, payload, pool)
	})

	t.Run("download object", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		var cmd client.PrmObjectGet

		hdr, read, err := pool.ObjectGetInit(ctxTimeout, containerID, objectID, signer, cmd)
		defer func() {
			_ = read.Close()
		}()

		require.NoError(t, err)
		require.NotNil(t, hdr.OwnerID())
		require.True(t, hdr.OwnerID().Equals(account))

		downloadedPayload := make([]byte, len(payload))

		l, err := read.Read(downloadedPayload)
		require.NoError(t, err)
		require.Equal(t, l, len(payload))

		require.True(t, bytes.Equal(payload, downloadedPayload))
	})

	t.Run("delete object", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		testDeleteObject(ctxTimeout, t, signer, containerID, objectID, pool)
		cl, err := pool.sdkClient()

		require.NoError(t, err)
		require.NoError(t, isObjectDeleted(ctxTimeout, cl, containerID, objectID, signer))
	})

	times := int(opts.sessionExpirationDuration * 3)
	for i := 0; i < times; i++ {
		epoch, err := tickNewEpoch(ctx, pool)
		require.NoError(t, err)

		t.Run(fmt.Sprintf("upload object through epoch:%d", epoch), func(t *testing.T) {
			ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut*time.Duration(times))
			defer cancel()

			payload = append(payload, 0x01) // Make it different from the one above, otherwise OID will be the same and we can get "status: code = 2052 message = object already removed"
			objID := testObjectPutInit(ctxTimeout, t, account, containerID, signer, payload, pool)
			objectList = append(objectList, objID)
		})
	}

	for _, objID := range objectList {
		epoch, err := tickNewEpoch(ctx, pool)
		require.NoError(t, err)

		t.Run(fmt.Sprintf("delete object through epoch:%d", epoch), func(t *testing.T) {
			ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut*time.Duration(times))
			defer cancel()

			testDeleteObject(ctxTimeout, t, signer, containerID, objID, pool)

			cl, err := pool.sdkClient()
			require.NoError(t, err)

			require.NoError(t, isObjectDeleted(ctxTimeout, cl, containerID, objID, signer))
		})
	}

	t.Run("container delete", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		testDeleteContainer(ctxTimeout, t, signer, containerID, pool)
		cl, err := pool.sdkClient()

		require.NoError(t, err)
		require.NoError(t, isBucketDeleted(ctxTimeout, cl, containerID))
	})

	t.Run("container really deleted", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		var prm client.PrmContainerGet
		_, err = pool.ContainerGet(ctxTimeout, containerID, prm)
		require.ErrorIs(t, err, apistatus.ErrContainerNotFound)
	})
}

func testPoolWaiterWithAIO(t *testing.T, nodeAddr string) {
	ctx := context.Background()

	account, signer, cont := testData(t)
	var eaclTable eacl.Table

	pool, err := New(NewFlatNodeParams([]string{nodeAddr}), signer, DefaultOptions())
	require.NoError(t, err)
	require.NoError(t, pool.Dial(ctx))

	var containerID cid.ID
	var objectID oid.ID

	payload := make([]byte, 8)
	_, err = rand.Read(payload)
	require.NoError(t, err)

	defaultPoolingTimeout := 1 * time.Second
	wait := waiter.NewWaiter(pool, defaultPoolingTimeout)

	t.Run("create container", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		containerID = testCreateContainer(ctxTimeout, t, signer, cont, wait)
		eaclTable = testEaclTable(containerID)
	})

	t.Run("set eacl", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		testSetEacl(ctxTimeout, t, signer, eaclTable, wait)
	})

	t.Run("get eacl", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		testGetEacl(ctxTimeout, t, containerID, eaclTable, pool)
	})

	t.Run("upload object", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		rf := object.RequiredFields{
			Container: containerID,
			Owner:     account,
		}

		var hdr object.Object
		hdr.InitCreation(rf)

		var prm client.PrmObjectPutInit
		prm.SetCopiesNumber(1)

		w, err := pool.ObjectPutInit(ctxTimeout, hdr, signer, prm)
		require.NoError(t, err)

		_, err = w.Write(payload)
		require.NoError(t, err)

		err = w.Close()
		require.NoError(t, err)

		objectID = w.GetResult().StoredObjectID()
	})

	t.Run("download object", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		var cmd client.PrmObjectGet

		hdr, read, err := pool.ObjectGetInit(ctxTimeout, containerID, objectID, signer, cmd)
		defer func() {
			_ = read.Close()
		}()

		require.NoError(t, err)
		require.NotNil(t, hdr.OwnerID())
		require.True(t, hdr.OwnerID().Equals(account))

		downloadedPayload := make([]byte, len(payload))

		l, err := read.Read(downloadedPayload)
		require.NoError(t, err)
		require.Equal(t, l, len(payload))

		require.True(t, bytes.Equal(payload, downloadedPayload))
	})

	t.Run("delete object", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		testDeleteObject(ctxTimeout, t, signer, containerID, objectID, pool)
	})

	t.Run("container delete", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		testDeleteContainer(ctxTimeout, t, signer, containerID, wait)
	})

	t.Run("container really deleted", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		var prm client.PrmContainerGet
		_, err = pool.ContainerGet(ctxTimeout, containerID, prm)
		require.ErrorIs(t, err, apistatus.ErrContainerNotFound)
	})
}

func testClientWaiterWithAIO(t *testing.T, nodeAddr string) {
	ctx := context.Background()

	account, signer, cont := testData(t)
	var eaclTable eacl.Table

	var prmInit client.PrmInit

	cl, err := client.New(prmInit)
	if err != nil {
		panic(fmt.Errorf("new client: %w", err))
	}

	// connect to NeoFS gateway
	var prmDial client.PrmDial
	prmDial.SetServerURI(nodeAddr) // endpoint address
	prmDial.SetTimeout(15 * time.Second)
	prmDial.SetStreamTimeout(15 * time.Second)

	if err = cl.Dial(prmDial); err != nil {
		panic(fmt.Errorf("dial %v", err))
	}

	var containerID cid.ID
	var objectID oid.ID

	payload := make([]byte, 8)
	_, err = rand.Read(payload)
	require.NoError(t, err)

	defaultPoolingTimeout := 1 * time.Second
	wait := waiter.NewWaiter(cl, defaultPoolingTimeout)

	t.Run("create container", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		containerID = testCreateContainer(ctxTimeout, t, signer, cont, wait)
		eaclTable = testEaclTable(containerID)
	})

	t.Run("set eacl", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		testSetEacl(ctxTimeout, t, signer, eaclTable, wait)
	})

	t.Run("get eacl", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		testGetEacl(ctxTimeout, t, containerID, eaclTable, cl)
	})

	t.Run("upload object", func(t *testing.T) {
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
		defer cancel()

		rf := object.RequiredFields{
			Container: containerID,
			Owner:     account,
		}

		var hdr object.Object
		hdr.InitCreation(rf)

		var prm client.PrmObjectPutInit
		prm.SetCopiesNumber(1)
		prm.WithinSession(sess)

		w, err := cl.ObjectPutInit(ctxTimeout, hdr, signer, prm)
		require.NoError(t, err)

		_, err = w.Write(payload)
		require.NoError(t, err)

		err = w.Close()
		require.NoError(t, err)

		objectID = w.GetResult().StoredObjectID()
	})

	t.Run("download object", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		var cmd client.PrmObjectGet

		hdr, read, err := cl.ObjectGetInit(ctxTimeout, containerID, objectID, signer, cmd)
		defer func() {
			_ = read.Close()
		}()

		require.NoError(t, err)
		require.NotNil(t, hdr.OwnerID())
		require.True(t, hdr.OwnerID().Equals(account))

		downloadedPayload := make([]byte, len(payload))

		l, err := read.Read(downloadedPayload)
		require.NoError(t, err)
		require.Equal(t, l, len(payload))

		require.True(t, bytes.Equal(payload, downloadedPayload))
	})

	t.Run("delete object", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		testDeleteObject(ctxTimeout, t, signer, containerID, objectID, cl)
	})

	t.Run("container delete", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		testDeleteContainer(ctxTimeout, t, signer, containerID, wait)
	})

	t.Run("container really deleted", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeOut)
		defer cancel()

		var prm client.PrmContainerGet
		_, err = cl.ContainerGet(ctxTimeout, containerID, prm)
		require.ErrorIs(t, err, apistatus.ErrContainerNotFound)
	})
}

func testObjectPutInit(ctx context.Context, t *testing.T, account user.ID, containerID cid.ID, signer user.Signer, payload []byte, putter objectPutIniter) oid.ID {
	rf := object.RequiredFields{
		Container: containerID,
		Owner:     account,
	}

	var hdr object.Object
	hdr.InitCreation(rf)

	var prm client.PrmObjectPutInit
	prm.SetCopiesNumber(1)

	w, err := putter.ObjectPutInit(ctx, hdr, signer, prm)
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
	pp.AddReplicas(rd)

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
	require.True(t, eacl.EqualTables(table, newTable))
}

func isBucketCreated(ctx context.Context, c containerGetter, id cid.ID) error {
	t := time.NewTicker(tickInterval)
	defer t.Stop()

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

func isBucketDeleted(ctx context.Context, c containerGetter, id cid.ID) error {
	t := time.NewTicker(tickInterval)
	defer t.Stop()

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
	oldBinary, err := oldTable.Marshal()
	if err != nil {
		return fmt.Errorf("oldTable.Marshal %w", err)
	}

	t := time.NewTicker(tickInterval)
	defer t.Stop()

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

			newBinary, err := table.Marshal()
			if err != nil {
				return fmt.Errorf("table.Marshal %w", err)
			}

			if bytes.Equal(oldBinary, newBinary) {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func isObjectDeleted(ctx context.Context, c objectHeadGetter, id cid.ID, oid oid.ID, signer user.Signer) error {
	t := time.NewTicker(tickInterval)
	defer t.Stop()

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
