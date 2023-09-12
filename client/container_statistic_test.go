package client

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	mathRand "math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-api-go/v2/accounting"
	v2acl "github.com/nspcc-dev/neofs-api-go/v2/acl"
	v2container "github.com/nspcc-dev/neofs-api-go/v2/container"
	netmapv2 "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/reputation"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	eacltest "github.com/nspcc-dev/neofs-sdk-go/eacl/test"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	reputation2 "github.com/nspcc-dev/neofs-sdk-go/reputation"
	session2 "github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
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

func randRefsContainerID() *refs.ContainerID {
	var id refs.ContainerID
	id.SetValue(randBytes(sha256.Size))
	return &id
}

func randContainerID() *cid.ID {
	var refID refs.ContainerID
	refID.SetValue(randBytes(sha256.Size))

	var id cid.ID
	_ = id.ReadFromV2(refID)

	return &id
}

func randAccount(signer user.Signer) *user.ID {
	u := signer.UserID()

	return &u
}

func randOwner(signer user.Signer) *refs.OwnerID {
	acc := randAccount(signer)

	var u refs.OwnerID
	acc.WriteToV2(&u)

	return &u
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
	pp.AddReplicas(rd)
	cont.SetPlacementPolicy(pp)

	return cont
}

func testEaclTable(containerID cid.ID) eacl.Table {
	record := eacl.NewRecord(eacl.ActionAllow, acl.OpObjectPut, eacl.NewTargetWithRole(eacl.RoleOthers))
	return eacl.NewForContainer(containerID, []eacl.Record{record})
}

func TestClientStatistic_AccountBalance(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPIBalance = func(cli *client.Client, req *accounting.BalanceRequest, opts ...client.CallOption) (*accounting.BalanceResponse, error) {
		var resp accounting.BalanceResponse
		var meta session.ResponseMetaHeader
		var balance accounting.Decimal
		var body accounting.BalanceResponseBody

		body.SetBalance(&balance)

		resp.SetBody(&body)
		resp.SetMetaHeader(&meta)

		err := signServiceMessage(signer, &resp)
		if err != nil {
			panic(fmt.Sprintf("sign response: %v", err))
		}

		return &resp, nil
	}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmBalanceGet
	prm.SetAccount(*randAccount(signer))
	_, err := c.BalanceGet(ctx, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodBalanceGet].requests)
}

func TestClientStatistic_ContainerPut(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPIPutContainer = func(cli *client.Client, req *v2container.PutRequest, opts ...client.CallOption) (*v2container.PutResponse, error) {
		var resp v2container.PutResponse
		var meta session.ResponseMetaHeader
		var body v2container.PutResponseBody

		body.SetContainerID(randRefsContainerID())

		resp.SetBody(&body)
		resp.SetMetaHeader(&meta)

		err := signServiceMessage(signer, &resp)
		if err != nil {
			panic(fmt.Sprintf("sign response: %v", err))
		}

		return &resp, nil
	}

	cont := prepareContainer(*randAccount(signer))

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmContainerPut
	_, err := c.ContainerPut(ctx, cont, signer, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodContainerPut].requests)
}

func TestClientStatistic_ContainerGet(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPIGetContainer = func(cli *client.Client, req *v2container.GetRequest, opts ...client.CallOption) (*v2container.GetResponse, error) {
		var cont v2container.Container
		var ver refs.Version
		var placementPolicyV2 netmapv2.PlacementPolicy
		var replicas []netmapv2.Replica
		var resp v2container.GetResponse
		var meta session.ResponseMetaHeader

		cont.SetOwnerID(randOwner(signer))
		cont.SetVersion(&ver)

		nonce, err := uuid.New().MarshalBinary()
		require.NoError(t, err)
		cont.SetNonce(nonce)

		replica := netmapv2.Replica{}
		replica.SetCount(1)
		replicas = append(replicas, replica)
		placementPolicyV2.SetReplicas(replicas)
		cont.SetPlacementPolicy(&placementPolicyV2)

		body := v2container.GetResponseBody{}
		body.SetContainer(&cont)

		resp.SetBody(&body)
		resp.SetMetaHeader(&meta)

		if err = signServiceMessage(signer, &resp); err != nil {
			panic(fmt.Sprintf("sign response: %v", err))
		}

		return &resp, nil
	}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmContainerGet
	_, err := c.ContainerGet(ctx, cid.ID{}, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodContainerGet].requests)
}

func TestClientStatistic_ContainerList(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPIListContainers = func(cli *client.Client, req *v2container.ListRequest, opts ...client.CallOption) (*v2container.ListResponse, error) {
		var resp v2container.ListResponse
		var meta session.ResponseMetaHeader
		var body v2container.ListResponseBody

		resp.SetBody(&body)
		resp.SetMetaHeader(&meta)

		if err := signServiceMessage(signer, &resp); err != nil {
			panic(fmt.Sprintf("sign response: %v", err))
		}

		return &resp, nil
	}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmContainerList
	_, err := c.ContainerList(ctx, *randAccount(signer), prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodContainerList].requests)
}

func TestClientStatistic_ContainerDelete(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPIDeleteContainer = func(cli *client.Client, req *v2container.DeleteRequest, opts ...client.CallOption) (*v2container.PutResponse, error) {
		var resp v2container.PutResponse
		var meta session.ResponseMetaHeader
		var body v2container.PutResponseBody

		resp.SetBody(&body)
		resp.SetMetaHeader(&meta)

		if err := signServiceMessage(signer, &resp); err != nil {
			panic(fmt.Sprintf("sign response: %v", err))
		}

		return &resp, nil
	}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmContainerDelete
	err := c.ContainerDelete(ctx, cid.ID{}, signer, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodContainerDelete].requests)
}

func TestClientStatistic_ContainerEacl(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPIGetEACL = func(cli *client.Client, req *v2container.GetExtendedACLRequest, opts ...client.CallOption) (*v2container.GetExtendedACLResponse, error) {
		var resp v2container.GetExtendedACLResponse
		var meta session.ResponseMetaHeader
		var aclTable v2acl.Table
		var body v2container.GetExtendedACLResponseBody

		eacltest.Table(t).WriteToV2(&aclTable)

		body.SetEACL(&aclTable)

		resp.SetBody(&body)
		resp.SetMetaHeader(&meta)

		if err := signServiceMessage(signer, &resp); err != nil {
			panic(fmt.Sprintf("sign response: %v", err))
		}

		return &resp, nil
	}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmContainerEACL
	_, err := c.ContainerEACL(ctx, cid.ID{}, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodContainerEACL].requests)
}

func TestClientStatistic_ContainerSetEacl(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPISetEACL = func(cli *client.Client, req *v2container.SetExtendedACLRequest, opts ...client.CallOption) (*v2container.PutResponse, error) {
		var resp v2container.PutResponse
		var meta session.ResponseMetaHeader
		var body v2container.PutResponseBody

		resp.SetBody(&body)
		resp.SetMetaHeader(&meta)

		if err := signServiceMessage(signer, &resp); err != nil {
			panic(fmt.Sprintf("sign response: %v", err))
		}

		return &resp, nil
	}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmContainerSetEACL
	table := testEaclTable(cid.ID{})
	err := c.ContainerSetEACL(ctx, table, signer, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodContainerSetEACL].requests)
}

func TestClientStatistic_ContainerAnnounceUsedSpace(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPIAnnounceUsedSpace = func(cli *client.Client, req *v2container.AnnounceUsedSpaceRequest, opts ...client.CallOption) (*v2container.PutResponse, error) {
		var resp v2container.PutResponse
		var meta session.ResponseMetaHeader
		var body v2container.PutResponseBody

		resp.SetBody(&body)
		resp.SetMetaHeader(&meta)

		if err := signServiceMessage(signer, &resp); err != nil {
			panic(fmt.Sprintf("sign response: %v", err))
		}

		return &resp, nil
	}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	estimation := container.SizeEstimation{}
	estimation.SetContainer(*randContainerID())
	estimation.SetValue(mathRand.Uint64())
	estimation.SetEpoch(mathRand.Uint64())

	var prm PrmAnnounceSpace
	err := c.ContainerAnnounceUsedSpace(ctx, []container.SizeEstimation{estimation}, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodContainerAnnounceUsedSpace].requests)
}

func TestClientStatistic_ContainerSyncContainerWithNetwork(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPINetworkInfo = func(cli *client.Client, req *netmapv2.NetworkInfoRequest, opts ...client.CallOption) (*netmapv2.NetworkInfoResponse, error) {
		var resp netmapv2.NetworkInfoResponse
		var meta session.ResponseMetaHeader
		var netInfo netmapv2.NetworkInfo
		var netConfig netmapv2.NetworkConfig
		var p1 netmapv2.NetworkParameter

		p1.SetKey(randBytes(10))
		p1.SetValue(randBytes(10))

		netConfig.SetParameters(p1)
		netInfo.SetNetworkConfig(&netConfig)

		body := netmapv2.NetworkInfoResponseBody{}
		body.SetNetworkInfo(&netInfo)

		resp.SetBody(&body)
		resp.SetMetaHeader(&meta)

		if err := signServiceMessage(signer, &resp); err != nil {
			panic(fmt.Sprintf("sign response: %v", err))
		}

		return &resp, nil
	}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	cont := prepareContainer(*randAccount(signer))

	err := SyncContainerWithNetwork(ctx, &cont, c)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodNetworkInfo].requests)
}

func TestClientStatistic_ContainerEndpointInfo(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPILocalNodeInfo = func(cli *client.Client, req *netmapv2.LocalNodeInfoRequest, opts ...client.CallOption) (*netmapv2.LocalNodeInfoResponse, error) {
		var resp netmapv2.LocalNodeInfoResponse
		var meta session.ResponseMetaHeader
		var ver refs.Version
		var nodeInfo netmapv2.NodeInfo

		nodeInfo.SetPublicKey(neofscrypto.PublicKeyBytes(signer.Public()))
		nodeInfo.SetAddresses("https://some-endpont.com")

		body := netmapv2.LocalNodeInfoResponseBody{}
		body.SetVersion(&ver)
		body.SetNodeInfo(&nodeInfo)

		resp.SetBody(&body)
		resp.SetMetaHeader(&meta)

		if err := signServiceMessage(signer, &resp); err != nil {
			panic(fmt.Sprintf("sign response: %v", err))
		}

		return &resp, nil
	}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	_, err := c.EndpointInfo(ctx, PrmEndpointInfo{})
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodEndpointInfo].requests)
}

func TestClientStatistic_ContainerNetMapSnapshot(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPINetMapSnapshot = func(cli *client.Client, req *netmapv2.SnapshotRequest, opts ...client.CallOption) (*netmapv2.SnapshotResponse, error) {
		var resp netmapv2.SnapshotResponse
		var meta session.ResponseMetaHeader
		var netMap netmapv2.NetMap

		body := netmapv2.SnapshotResponseBody{}
		body.SetNetMap(&netMap)

		resp.SetBody(&body)
		resp.SetMetaHeader(&meta)

		if err := signServiceMessage(signer, &resp); err != nil {
			panic(fmt.Sprintf("sign response: %v", err))
		}

		return &resp, nil
	}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect
	c.setNeoFSAPIServer((*coreServer)(&c.c))

	_, err := c.NetMapSnapshot(ctx, PrmNetMapSnapshot{})
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodNetMapSnapshot].requests)
}

func TestClientStatistic_CreateSession(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPICreateSession = func(cli *client.Client, req *session.CreateRequest, opts ...client.CallOption) (*session.CreateResponse, error) {
		var resp session.CreateResponse
		var meta session.ResponseMetaHeader

		body := session.CreateResponseBody{}
		body.SetID(randBytes(10))

		body.SetSessionKey(neofscrypto.PublicKeyBytes(signer.Public()))

		resp.SetBody(&body)
		resp.SetMetaHeader(&meta)

		if err := signServiceMessage(signer, &resp); err != nil {
			panic(fmt.Sprintf("sign response: %v", err))
		}

		return &resp, nil
	}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect
	c.setNeoFSAPIServer((*coreServer)(&c.c))

	var prm PrmSessionCreate

	_, err := c.SessionCreate(ctx, signer, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodSessionCreate].requests)
}

func TestClientStatistic_ObjectPut(t *testing.T) {
	t.Skip("need changes to api-go, to set `wc client.MessageWriterCloser` in rpcapi.PutRequestWriter")

	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPIPutObject = func(cli *client.Client, pResp *v2object.PutResponse, opts ...client.CallOption) (objectWriter, error) {
		var resp rpcapi.PutRequestWriter

		return &resp, nil
	}

	containerID := *randContainerID()
	account := randAccount(signer)

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect
	c.setNeoFSAPIServer((*coreServer)(&c.c))

	var tokenSession session2.Object
	tokenSession.SetID(uuid.New())
	tokenSession.SetExp(1)
	tokenSession.BindContainer(containerID)
	tokenSession.ForVerb(session2.VerbObjectPut)
	tokenSession.SetAuthKey(signer.Public())
	tokenSession.SetIssuer(*account)

	err := tokenSession.Sign(signer)
	require.NoError(t, err)

	var prm PrmObjectPutInit
	prm.WithinSession(tokenSession)

	var hdr object.Object
	hdr.SetOwnerID(account)
	hdr.SetContainerID(containerID)

	writer, err := c.ObjectPutInit(ctx, hdr, signer, prm)
	require.NoError(t, err)

	_, err = writer.Write(randBytes(10))
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	require.Equal(t, 2, collector.methods[stat.MethodObjectPut].requests)
}

func TestClientStatistic_ObjectDelete(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPIDeleteObject = func(cli *client.Client, req *v2object.DeleteRequest, opts ...client.CallOption) (*v2object.DeleteResponse, error) {
		var resp v2object.DeleteResponse
		var meta session.ResponseMetaHeader
		var body v2object.DeleteResponseBody
		var addr refs.Address
		var objID refs.ObjectID
		var contID = randRefsContainerID()

		objID.SetValue(randBytes(32))

		addr.SetContainerID(contID)
		addr.SetObjectID(&objID)

		body.SetTombstone(&addr)

		resp.SetBody(&body)
		resp.SetMetaHeader(&meta)

		if err := signServiceMessage(signer, &resp); err != nil {
			panic(fmt.Sprintf("sign response: %v", err))
		}

		return &resp, nil
	}

	containerID := *randContainerID()
	objectID := oid.ID{}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmObjectDelete

	_, err := c.ObjectDelete(ctx, containerID, objectID, signer, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodObjectDelete].requests)
}

func TestClientStatistic_ObjectGet(t *testing.T) {
	t.Skip("need changes to api-go, to set `r client.MessageReader` in rpcapi.GetResponseReader")

	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPIGetObject = func(cli *client.Client, req *v2object.GetRequest, opts ...client.CallOption) (*rpcapi.GetResponseReader, error) {
		var resp rpcapi.GetResponseReader

		// todo: fill

		return &resp, nil
	}

	containerID := *randContainerID()
	objectID := oid.ID{}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmObjectGet

	_, reader, err := c.ObjectGetInit(ctx, containerID, objectID, signer, prm)
	require.NoError(t, err)

	buff := make([]byte, 32)
	_, err = reader.Read(buff)
	require.NoError(t, err)

	require.Equal(t, 2, collector.methods[stat.MethodObjectGet].requests)
}

func TestClientStatistic_ObjectHead(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPIHeadObject = func(cli *client.Client, req *v2object.HeadRequest, opts ...client.CallOption) (*v2object.HeadResponse, error) {
		var resp v2object.HeadResponse
		var meta session.ResponseMetaHeader
		var body v2object.HeadResponseBody
		var headerPart v2object.HeaderWithSignature

		body.SetHeaderPart(&headerPart)

		resp.SetBody(&body)
		resp.SetMetaHeader(&meta)

		if err := signServiceMessage(signer, &resp); err != nil {
			panic(fmt.Sprintf("sign response: %v", err))
		}

		return &resp, nil
	}

	containerID := *randContainerID()
	objectID := oid.ID{}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmObjectHead

	_, err := c.ObjectHead(ctx, containerID, objectID, signer, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodObjectHead].requests)
}

func TestClientStatistic_ObjectRange(t *testing.T) {
	t.Skip("need changes to api-go, to set `r client.MessageReader` in rpcapi.ObjectRangeResponseReader")

	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPIGetObjectRange = func(cli *client.Client, req *v2object.GetRangeRequest, opts ...client.CallOption) (*rpcapi.ObjectRangeResponseReader, error) {
		var resp rpcapi.ObjectRangeResponseReader

		// todo: fill

		return &resp, nil
	}

	containerID := *randContainerID()
	objectID := oid.ID{}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmObjectRange

	reader, err := c.ObjectRangeInit(ctx, containerID, objectID, 0, 1, signer, prm)
	require.NoError(t, err)

	buff := make([]byte, 32)
	_, err = reader.Read(buff)
	require.NoError(t, err)

	require.Equal(t, 2, collector.methods[stat.MethodObjectRange].requests)
}

func TestClientStatistic_ObjectHash(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPIHashObjectRange = func(cli *client.Client, req *v2object.GetRangeHashRequest, opts ...client.CallOption) (*v2object.GetRangeHashResponse, error) {
		var resp v2object.GetRangeHashResponse
		var meta session.ResponseMetaHeader
		var body v2object.GetRangeHashResponseBody

		body.SetHashList([][]byte{
			randBytes(4),
		})

		resp.SetBody(&body)
		resp.SetMetaHeader(&meta)

		if err := signServiceMessage(signer, &resp); err != nil {
			panic(fmt.Sprintf("sign response: %v", err))
		}

		return &resp, nil
	}

	containerID := *randContainerID()
	objectID := oid.ID{}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmObjectHash
	prm.SetRangeList(0, 2)

	_, err := c.ObjectHash(ctx, containerID, objectID, signer, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodObjectHash].requests)
}

func TestClientStatistic_ObjectSearch(t *testing.T) {
	t.Skip("need changes to api-go, to set `r client.MessageReader` in rpcapi.SearchResponseReader")

	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPISearchObjects = func(cli *client.Client, req *v2object.SearchRequest, opts ...client.CallOption) (*rpcapi.SearchResponseReader, error) {
		var resp rpcapi.SearchResponseReader

		// todo: fill

		return &resp, nil
	}

	containerID := *randContainerID()

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var prm PrmObjectSearch

	reader, err := c.ObjectSearchInit(ctx, containerID, signer, prm)
	require.NoError(t, err)

	iterator := func(oid.ID) bool {
		return false
	}

	err = reader.Iterate(iterator)
	require.NoError(t, err)

	require.Equal(t, 2, collector.methods[stat.MethodObjectSearch].requests)
}

func TestClientStatistic_AnnounceIntermediateTrust(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPIAnnounceIntermediateResult = func(cli *client.Client, req *reputation.AnnounceIntermediateResultRequest, opts ...client.CallOption) (*reputation.AnnounceIntermediateResultResponse, error) {
		var resp reputation.AnnounceIntermediateResultResponse
		var meta session.ResponseMetaHeader
		var body reputation.AnnounceIntermediateResultResponseBody

		resp.SetBody(&body)
		resp.SetMetaHeader(&meta)

		if err := signServiceMessage(signer, &resp); err != nil {
			panic(fmt.Sprintf("sign response: %v", err))
		}

		return &resp, nil
	}

	collector := newCollector()
	c.prm.statisticCallback = collector.Collect

	var trust reputation2.PeerToPeerTrust
	var prm PrmAnnounceIntermediateTrust

	err := c.AnnounceIntermediateTrust(ctx, 1, trust, prm)
	require.NoError(t, err)

	require.Equal(t, 1, collector.methods[stat.MethodAnnounceIntermediateTrust].requests)
}

func TestClientStatistic_MethodAnnounceLocalTrust(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)
	ctx := context.Background()
	c := newClient(t, nil)

	rpcAPIAnnounceLocalTrust = func(cli *client.Client, req *reputation.AnnounceLocalTrustRequest, opts ...client.CallOption) (*reputation.AnnounceLocalTrustResponse, error) {
		var resp reputation.AnnounceLocalTrustResponse
		var meta session.ResponseMetaHeader
		var body reputation.AnnounceLocalTrustResponseBody

		resp.SetBody(&body)
		resp.SetMetaHeader(&meta)

		if err := signServiceMessage(signer, &resp); err != nil {
			panic(fmt.Sprintf("sign response: %v", err))
		}

		return &resp, nil
	}

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
