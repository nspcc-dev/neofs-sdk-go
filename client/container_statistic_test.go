package client

import (
	"context"
	"crypto/rand"
	"io"
	mathRand "math/rand/v2"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	reputation2 "github.com/nspcc-dev/neofs-sdk-go/reputation"
	session2 "github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

type (
	methodStatistic struct {
		requests int
		errors   int
		duration time.Duration
	}

	testStatCollector struct {
		methods map[stat.Method]*methodStatistic
	}
)

func newCollector() *testStatCollector {
	c := testStatCollector{
		methods: make(map[stat.Method]*methodStatistic),
	}

	for i := stat.MethodBalanceGet; i < stat.MethodLast; i++ {
		c.methods[i] = &methodStatistic{}
	}

	return &c
}

func (c *testStatCollector) Collect(_ []byte, _ string, method stat.Method, duration time.Duration, err error) {
	data, ok := c.methods[method]
	if ok {
		data.duration += duration
		if duration > 0 {
			data.requests++
		}

		if err != nil {
			data.errors++
		}
	}
}

func randBytes(l int) []byte {
	r := make([]byte, l)
	_, _ = rand.Read(r)

	return r
}

func prepareContainer(accountID user.ID) container.Container {
	cont := container.Container{}
	cont.Init()
	cont.SetOwner(accountID)
	cont.SetBasicACL(acl.PublicRW)

	cont.SetName(strconv.FormatInt(time.Now().UnixNano(), 16))
	cont.SetCreationTime(time.Now().UTC())

	var pp netmap.PlacementPolicy
	var rd netmap.ReplicaDescriptor
	rd.SetNumberOfObjects(1)

	pp.SetContainerBackupFactor(1)
	pp.SetReplicas([]netmap.ReplicaDescriptor{rd})
	cont.SetPlacementPolicy(pp)

	return cont
}

func testEaclTable(containerID cid.ID) eacl.Table {
	var table eacl.Table
	table.SetCID(containerID)

	r := eacl.ConstructRecord(eacl.ActionAllow, eacl.OperationPut, []eacl.Target{eacl.NewTargetByRole(eacl.RoleOthers)})
	table.AddRecord(&r)

	return table
}

func TestClientStatistic_ContainerPut(t *testing.T) {
	usr := usertest.User()
	ctx := context.Background()
	var srv testPutContainerServer
	c := newTestContainerClient(t, &srv)
	cont := prepareContainer(usr.ID)

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmContainerPut
	_, err := c.ContainerPut(ctx, cont, usr.RFC6979, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodContainerPut].requests)
}

func TestClientStatistic_ContainerGet(t *testing.T) {
	ctx := context.Background()
	var srv testGetContainerServer
	c := newTestContainerClient(t, &srv)
	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmContainerGet
	_, err := c.ContainerGet(ctx, cid.ID{}, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodContainerGet].requests)
}

func TestClientStatistic_ContainerList(t *testing.T) {
	usr := usertest.User()
	ctx := context.Background()
	var srv testListContainersServer
	c := newTestContainerClient(t, &srv)
	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmContainerList
	_, err := c.ContainerList(ctx, usr.ID, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodContainerList].requests)
}

func TestClientStatistic_ContainerDelete(t *testing.T) {
	usr := usertest.User()
	ctx := context.Background()
	var srv testDeleteContainerServer
	c := newTestContainerClient(t, &srv)
	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmContainerDelete
	err := c.ContainerDelete(ctx, cid.ID{}, usr, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodContainerDelete].requests)
}

func TestClientStatistic_ContainerEacl(t *testing.T) {
	ctx := context.Background()
	var srv testGetEACLServer
	c := newTestContainerClient(t, &srv)
	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmContainerEACL
	_, err := c.ContainerEACL(ctx, cid.ID{}, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodContainerEACL].requests)
}

func TestClientStatistic_ContainerSetEacl(t *testing.T) {
	usr := usertest.User()
	ctx := context.Background()
	var srv testSetEACLServer
	c := newTestContainerClient(t, &srv)
	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmContainerSetEACL
	table := testEaclTable(cidtest.ID())
	err := c.ContainerSetEACL(ctx, table, usr, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodContainerSetEACL].requests)
}

func TestClientStatistic_ContainerAnnounceUsedSpace(t *testing.T) {
	ctx := context.Background()
	var srv testAnnounceContainerSpaceServer
	c := newTestContainerClient(t, &srv)
	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	estimation := container.SizeEstimation{}
	estimation.SetContainer(cidtest.ID())
	estimation.SetValue(mathRand.Uint64())
	estimation.SetEpoch(mathRand.Uint64())

	var prm PrmAnnounceSpace
	err := c.ContainerAnnounceUsedSpace(ctx, []container.SizeEstimation{estimation}, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodContainerAnnounceUsedSpace].requests)
}

func TestClientStatistic_ContainerSyncContainerWithNetwork(t *testing.T) {
	usr := usertest.User()
	ctx := context.Background()
	var srv testGetNetworkInfoServer
	c := newTestNetmapClient(t, &srv)
	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	cont := prepareContainer(usr.ID)

	err := SyncContainerWithNetwork(ctx, &cont, c)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodNetworkInfo].requests)
}

func TestClientStatistic_ContainerEndpointInfo(t *testing.T) {
	ctx := context.Background()
	srv := newTestGetNodeInfoServer()
	c := newTestNetmapClient(t, srv)
	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	_, err := c.EndpointInfo(ctx, PrmEndpointInfo{})
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodEndpointInfo].requests)
}

func TestClientStatistic_ContainerNetMapSnapshot(t *testing.T) {
	ctx := context.Background()
	var srv testNetmapSnapshotServer
	c := newTestNetmapClient(t, &srv)
	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	_, err := c.NetMapSnapshot(ctx, PrmNetMapSnapshot{})
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodNetMapSnapshot].requests)
}

func TestClientStatistic_CreateSession(t *testing.T) {
	usr := usertest.User()
	ctx := context.Background()
	var srv testCreateSessionServer
	c := newTestSessionClient(t, &srv)
	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmSessionCreate

	_, err := c.SessionCreate(ctx, usr, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodSessionCreate].requests)
}

func TestClientStatistic_ObjectPut(t *testing.T) {
	usr := usertest.User()
	ctx := context.Background()
	var srv testPutObjectServer
	c := newTestObjectClient(t, &srv)
	containerID := cidtest.ID()

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var tokenSession session2.Object
	tokenSession.SetID(uuid.New())
	tokenSession.SetExp(1)
	tokenSession.BindContainer(containerID)
	tokenSession.ForVerb(session2.VerbObjectPut)
	tokenSession.SetAuthKey(usr.Public())
	tokenSession.SetIssuer(usr.ID)

	err := tokenSession.Sign(usr)
	require.NoError(t, err)

	var prm PrmObjectPutInit
	prm.WithinSession(tokenSession)

	var hdr object.Object
	hdr.SetOwner(usr.ID)
	hdr.SetContainerID(containerID)

	writer, err := c.ObjectPutInit(ctx, hdr, usr, prm)
	require.NoError(t, err)

	_, err = writer.Write(randBytes(10))
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodObjectPut].requests)
	require.Equal(t, 1, collector.methods[stat.MethodObjectPutStream].requests)
}

func TestClientStatistic_ObjectDelete(t *testing.T) {
	usr := usertest.User()
	ctx := context.Background()
	var srv testDeleteObjectServer
	c := newTestObjectClient(t, &srv)
	containerID := cidtest.ID()
	objectID := oid.ID{}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmObjectDelete

	_, err := c.ObjectDelete(ctx, containerID, objectID, usr, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodObjectDelete].requests)
}

func TestClientStatistic_ObjectGet(t *testing.T) {
	usr := usertest.User()
	ctx := context.Background()
	var srv testGetObjectServer
	c := newTestObjectClient(t, &srv)
	containerID := cidtest.ID()
	objectID := oid.ID{}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmObjectGet

	_, reader, err := c.ObjectGetInit(ctx, containerID, objectID, usr, prm)
	require.NoError(t, err)
	_, err = io.Copy(io.Discard, reader)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodObjectGet].requests)
}

func TestClientStatistic_ObjectHead(t *testing.T) {
	usr := usertest.User()
	ctx := context.Background()
	var srv testHeadObjectServer
	c := newTestObjectClient(t, &srv)
	containerID := cidtest.ID()
	objectID := oid.ID{}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmObjectHead

	_, err := c.ObjectHead(ctx, containerID, objectID, usr, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodObjectHead].requests)
}

func TestClientStatistic_ObjectRange(t *testing.T) {
	usr := usertest.User()
	ctx := context.Background()
	var srv testGetObjectPayloadRangeServer
	c := newTestObjectClient(t, &srv)
	containerID := cidtest.ID()
	objectID := oid.ID{}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmObjectRange

	reader, err := c.ObjectRangeInit(ctx, containerID, objectID, 0, 1, usr, prm)
	require.NoError(t, err)
	_, err = io.Copy(io.Discard, reader)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodObjectRange].requests)
}

func TestClientStatistic_ObjectHash(t *testing.T) {
	usr := usertest.User()
	ctx := context.Background()
	var srv testHashObjectPayloadRangesServer
	c := newTestObjectClient(t, &srv)
	containerID := cidtest.ID()
	objectID := oid.ID{}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmObjectHash
	prm.SetRangeList(0, 2)

	_, err := c.ObjectHash(ctx, containerID, objectID, usr, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodObjectHash].requests)
}

func TestClientStatistic_ObjectSearch(t *testing.T) {
	usr := usertest.User()
	ctx := context.Background()
	var srv testSearchObjectsServer
	c := newTestObjectClient(t, &srv)
	containerID := cidtest.ID()

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmObjectSearch

	reader, err := c.ObjectSearchInit(ctx, containerID, usr, prm)
	require.NoError(t, err)

	iterator := func(oid.ID) bool {
		return false
	}

	err = reader.Iterate(iterator)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodObjectSearch].requests)
}

func TestClientStatistic_AnnounceIntermediateTrust(t *testing.T) {
	ctx := context.Background()
	var srv testAnnounceIntermediateReputationServer
	c := newTestReputationClient(t, &srv)
	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var trust reputation2.PeerToPeerTrust
	var prm PrmAnnounceIntermediateTrust

	err := c.AnnounceIntermediateTrust(ctx, 1, trust, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodAnnounceIntermediateTrust].requests)
}

func TestClientStatistic_MethodAnnounceLocalTrust(t *testing.T) {
	ctx := context.Background()
	var srv testAnnounceLocalTrustServer
	c := newTestReputationClient(t, &srv)
	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var peer reputation2.PeerID
	var trust reputation2.Trust
	trust.SetPeer(peer)

	var prm PrmAnnounceLocalTrust

	err := c.AnnounceLocalTrust(ctx, 1, []reputation2.Trust{trust}, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodAnnounceLocalTrust].requests)
}
