package audit_test

import (
	"bytes"
	"testing"

	apiaudit "github.com/nspcc-dev/neofs-sdk-go/api/audit"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/nspcc-dev/neofs-sdk-go/audit"
	audittest "github.com/nspcc-dev/neofs-sdk-go/audit/test"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestResult_Version(t *testing.T) {
	var r audit.Result
	r.ForContainer(cidtest.ID()) // just to satisfy Unmarshal
	var msg1, msg2 apiaudit.DataAuditResult

	b := r.Marshal()
	err := proto.Unmarshal(b, &msg1)
	require.NoError(t, err)
	require.Equal(t, &refs.Version{Major: 2, Minor: 13}, msg1.Version)

	msg2.Version = &refs.Version{Major: 3, Minor: 14}
	msg2.ContainerId = &refs.ContainerID{Value: make([]byte, 32)} // just to satisfy Unmarshal
	b, err = proto.Marshal(&msg2)
	require.NoError(t, err)
	err = r.Unmarshal(b)
	require.NoError(t, err)
	err = proto.Unmarshal(r.Marshal(), &msg1)
	require.Equal(t, &refs.Version{Major: 3, Minor: 14}, msg1.Version)
}

func TestResult_Marshal(t *testing.T) {
	r := audittest.Result()

	data := r.Marshal()

	var r2 audit.Result
	require.NoError(t, r2.Unmarshal(data))

	require.Equal(t, r, r2)

	t.Run("invalid protobuf", func(t *testing.T) {
		var r audit.Result
		msg := []byte("definitely_not_protobuf")
		err := r.Unmarshal(msg)
		require.ErrorContains(t, err, "decode protobuf")
	})
	t.Run("missing container", func(t *testing.T) {
		var msg apiaudit.DataAuditResult
		require.NoError(t, proto.Unmarshal(r.Marshal(), &msg))
		msg.ContainerId = nil
		b, err := proto.Marshal(&msg)
		require.NoError(t, err)

		err = r.Unmarshal(b)
		require.ErrorContains(t, err, "container ID is not set")
	})
	t.Run("invalid container", func(t *testing.T) {
		var msg apiaudit.DataAuditResult
		require.NoError(t, proto.Unmarshal(r.Marshal(), &msg))
		msg.ContainerId = &refs.ContainerID{Value: []byte("invalid_container")}
		b, err := proto.Marshal(&msg)
		require.NoError(t, err)

		err = r.Unmarshal(b)
		require.ErrorContains(t, err, "invalid container")
	})
	t.Run("invalid passed SG", func(t *testing.T) {
		r := r
		r.SetPassedStorageGroups([]oid.ID{oidtest.ID(), oidtest.ID()})
		var msg apiaudit.DataAuditResult
		require.NoError(t, proto.Unmarshal(r.Marshal(), &msg))
		msg.PassSg[1].Value = []byte("invalid_object")
		b, err := proto.Marshal(&msg)
		require.NoError(t, err)

		err = r.Unmarshal(b)
		require.ErrorContains(t, err, "invalid passed storage group ID #1")
	})
	t.Run("invalid failed SG", func(t *testing.T) {
		r := r
		r.SetFailedStorageGroups([]oid.ID{oidtest.ID(), oidtest.ID()})
		var msg apiaudit.DataAuditResult
		require.NoError(t, proto.Unmarshal(r.Marshal(), &msg))
		msg.FailSg[1].Value = []byte("invalid_object")
		b, err := proto.Marshal(&msg)
		require.NoError(t, err)

		err = r.Unmarshal(b)
		require.ErrorContains(t, err, "invalid failed storage group ID #1")
	})
}

func TestResult_Epoch(t *testing.T) {
	var r audit.Result
	r.ForContainer(cidtest.ID()) // just to satisfy Unmarshal
	var msg apiaudit.DataAuditResult

	require.Zero(t, r.Epoch())

	r.ForEpoch(42)
	require.EqualValues(t, 42, r.Epoch())

	b := r.Marshal()
	err := proto.Unmarshal(b, &msg)
	require.NoError(t, err)

	b, err = proto.Marshal(&msg)
	require.NoError(t, err)

	r.ForEpoch(43) // any other
	require.EqualValues(t, 43, r.Epoch())

	err = r.Unmarshal(b)
	require.NoError(t, err)
	require.EqualValues(t, 42, r.Epoch())
}

func TestResult_Container(t *testing.T) {
	var r audit.Result
	var msg apiaudit.DataAuditResult
	cnr := cidtest.ID()

	_, ok := r.Container()
	require.False(t, ok)

	r.ForContainer(cnr)
	res, ok := r.Container()
	require.True(t, ok)
	require.Equal(t, cnr, res)

	b := r.Marshal()
	err := proto.Unmarshal(b, &msg)
	require.NoError(t, err)

	b, err = proto.Marshal(&msg)
	require.NoError(t, err)

	cnrOther := cidtest.ChangeID(cnr)
	r.ForContainer(cnrOther)
	require.True(t, ok)
	require.Equal(t, cnr, res)

	err = r.Unmarshal(b)
	require.NoError(t, err)
	res, ok = r.Container()
	require.True(t, ok)
	require.Equal(t, cnr, res)
}

func TestResult_AuditorKey(t *testing.T) {
	var r audit.Result
	r.ForContainer(cidtest.ID()) // just to satisfy Unmarshal
	var msg apiaudit.DataAuditResult
	key := []byte("any_key")

	require.Zero(t, r.AuditorKey())

	b := r.Marshal()
	err := r.Unmarshal(b)
	require.NoError(t, err)
	require.Nil(t, r.AuditorKey())

	r.SetAuditorKey(key)
	require.Equal(t, key, r.AuditorKey())

	b = r.Marshal()
	err = proto.Unmarshal(b, &msg)
	require.NoError(t, err)

	b, err = proto.Marshal(&msg)
	require.NoError(t, err)

	keyOther := bytes.Clone(key)
	keyOther[0]++
	r.SetAuditorKey(keyOther)
	require.Equal(t, keyOther, r.AuditorKey())

	err = r.Unmarshal(b)
	require.NoError(t, err)
	require.Equal(t, key, r.AuditorKey())
}

func TestResult_Completed(t *testing.T) {
	var r audit.Result
	r.ForContainer(cidtest.ID()) // just to satisfy Unmarshal
	var msg apiaudit.DataAuditResult

	require.Zero(t, r.Completed())

	b := r.Marshal()
	err := r.Unmarshal(b)
	require.NoError(t, err)
	require.False(t, r.Completed())

	r.SetCompleted(true)
	require.True(t, r.Completed())

	b = r.Marshal()
	err = proto.Unmarshal(b, &msg)
	require.NoError(t, err)

	b, err = proto.Marshal(&msg)
	require.NoError(t, err)

	r.SetCompleted(false)
	require.False(t, r.Completed())

	err = r.Unmarshal(b)
	require.NoError(t, err)
	require.True(t, r.Completed())
}

func TestResult_RequestsPoR(t *testing.T) {
	var r audit.Result
	r.ForContainer(cidtest.ID()) // just to satisfy Unmarshal
	var msg apiaudit.DataAuditResult
	const val = 64304

	require.Zero(t, r.RequestsPoR())

	b := r.Marshal()
	err := r.Unmarshal(b)
	require.NoError(t, err)
	require.Zero(t, r.RequestsPoR())

	r.SetRequestsPoR(val)
	require.EqualValues(t, val, r.RequestsPoR())

	b = r.Marshal()
	err = proto.Unmarshal(b, &msg)
	require.NoError(t, err)

	b, err = proto.Marshal(&msg)
	require.NoError(t, err)

	valOther := val + 1
	r.SetRequestsPoR(uint32(valOther))
	require.EqualValues(t, valOther, r.RequestsPoR())

	err = r.Unmarshal(b)
	require.NoError(t, err)
	require.EqualValues(t, val, r.RequestsPoR())
}

func TestResult_RetriesPoR(t *testing.T) {
	var r audit.Result
	r.ForContainer(cidtest.ID()) // just to satisfy Unmarshal
	var msg apiaudit.DataAuditResult
	const val = 984609

	require.Zero(t, r.RetriesPoR())

	b := r.Marshal()
	err := r.Unmarshal(b)
	require.NoError(t, err)
	require.Zero(t, r.RetriesPoR())

	r.SetRetriesPoR(val)
	require.EqualValues(t, val, r.RetriesPoR())

	b = r.Marshal()
	err = proto.Unmarshal(b, &msg)
	require.NoError(t, err)

	b, err = proto.Marshal(&msg)
	require.NoError(t, err)

	valOther := val + 1
	r.SetRetriesPoR(uint32(valOther))
	require.EqualValues(t, valOther, r.RetriesPoR())

	err = r.Unmarshal(b)
	require.NoError(t, err)
	require.EqualValues(t, val, r.RetriesPoR())
}

func TestResult_Hits(t *testing.T) {
	var r audit.Result
	r.ForContainer(cidtest.ID()) // just to satisfy Unmarshal
	var msg apiaudit.DataAuditResult
	const val = 23641

	require.Zero(t, r.Hits())

	b := r.Marshal()
	err := r.Unmarshal(b)
	require.NoError(t, err)
	require.Zero(t, r.Hits())

	r.SetHits(val)
	require.EqualValues(t, val, r.Hits())

	b = r.Marshal()
	err = proto.Unmarshal(b, &msg)
	require.NoError(t, err)

	b, err = proto.Marshal(&msg)
	require.NoError(t, err)

	valOther := val + 1
	r.SetHits(uint32(valOther))
	require.EqualValues(t, valOther, r.Hits())

	err = r.Unmarshal(b)
	require.NoError(t, err)
	require.EqualValues(t, val, r.Hits())
}

func TestResult_Misses(t *testing.T) {
	var r audit.Result
	r.ForContainer(cidtest.ID()) // just to satisfy Unmarshal
	var msg apiaudit.DataAuditResult
	const val = 684975

	require.Zero(t, r.Misses())

	b := r.Marshal()
	err := r.Unmarshal(b)
	require.NoError(t, err)
	require.Zero(t, r.Misses())

	r.SetMisses(val)
	require.EqualValues(t, val, r.Misses())

	b = r.Marshal()
	err = proto.Unmarshal(b, &msg)
	require.NoError(t, err)

	b, err = proto.Marshal(&msg)
	require.NoError(t, err)

	valOther := val + 1
	r.SetMisses(uint32(valOther))
	require.EqualValues(t, valOther, r.Misses())

	err = r.Unmarshal(b)
	require.NoError(t, err)
	require.EqualValues(t, val, r.Misses())
}

func TestResult_Failures(t *testing.T) {
	var r audit.Result
	r.ForContainer(cidtest.ID()) // just to satisfy Unmarshal
	var msg apiaudit.DataAuditResult
	const val = 25927509

	require.Zero(t, r.Failures())

	b := r.Marshal()
	err := r.Unmarshal(b)
	require.NoError(t, err)
	require.Zero(t, r.Failures())

	r.SetFailures(val)
	require.EqualValues(t, val, r.Failures())

	b = r.Marshal()
	err = proto.Unmarshal(b, &msg)
	require.NoError(t, err)

	b, err = proto.Marshal(&msg)
	require.NoError(t, err)
	valOther := val + 1
	r.SetFailures(uint32(valOther))
	require.EqualValues(t, valOther, r.Failures())

	err = r.Unmarshal(b)
	require.NoError(t, err)
	require.EqualValues(t, val, r.Failures())
}

func TestResult_PassedStorageGroups(t *testing.T) {
	var r audit.Result
	r.ForContainer(cidtest.ID()) // just to satisfy Unmarshal
	var msg apiaudit.DataAuditResult
	ids := []oid.ID{oidtest.ID(), oidtest.ID()}

	require.Zero(t, r.PassedStorageGroups())

	b := r.Marshal()
	err := r.Unmarshal(b)
	require.NoError(t, err)
	require.Zero(t, r.PassedStorageNodes())

	r.SetPassedStorageGroups(ids)
	require.Equal(t, ids, r.PassedStorageGroups())

	b = r.Marshal()
	err = proto.Unmarshal(b, &msg)
	require.NoError(t, err)

	b, err = proto.Marshal(&msg)
	require.NoError(t, err)

	idsOther := make([]oid.ID, len(ids))
	for i := range idsOther {
		idsOther[i] = oidtest.ChangeID(ids[i])
	}

	r.SetPassedStorageGroups(idsOther)
	require.Equal(t, idsOther, r.PassedStorageGroups())

	err = r.Unmarshal(b)
	require.NoError(t, err)
	require.Equal(t, ids, r.PassedStorageGroups())
}

func TestResult_FailedStorageGroups(t *testing.T) {
	var r audit.Result
	r.ForContainer(cidtest.ID()) // just to satisfy Unmarshal
	var msg apiaudit.DataAuditResult
	ids := []oid.ID{oidtest.ID(), oidtest.ID()}

	require.Zero(t, r.FailedStorageGroups())

	b := r.Marshal()
	err := r.Unmarshal(b)
	require.NoError(t, err)
	require.Zero(t, r.FailedStorageGroups())

	r.SetFailedStorageGroups(ids)
	require.Equal(t, ids, r.FailedStorageGroups())

	b = r.Marshal()
	err = proto.Unmarshal(b, &msg)
	require.NoError(t, err)

	b, err = proto.Marshal(&msg)
	require.NoError(t, err)

	idsOther := make([]oid.ID, len(ids))
	for i := range idsOther {
		idsOther[i] = oidtest.ChangeID(ids[i])
	}

	r.SetFailedStorageGroups(idsOther)
	require.Equal(t, idsOther, r.FailedStorageGroups())

	err = r.Unmarshal(b)
	require.NoError(t, err)
	require.Equal(t, ids, r.FailedStorageGroups())
}

func TestResult_PassedStorageNodes(t *testing.T) {
	var r audit.Result
	r.ForContainer(cidtest.ID()) // just to satisfy Unmarshal
	var msg apiaudit.DataAuditResult
	keys := [][]byte{
		[]byte("any_key1"),
		[]byte("any_key2"),
	}

	require.Zero(t, r.PassedStorageNodes())

	b := r.Marshal()
	err := r.Unmarshal(b)
	require.NoError(t, err)
	require.Zero(t, r.PassedStorageNodes())

	r.SetPassedStorageNodes(keys)
	require.Equal(t, keys, r.PassedStorageNodes())

	b = r.Marshal()
	err = proto.Unmarshal(b, &msg)
	require.NoError(t, err)

	b, err = proto.Marshal(&msg)
	require.NoError(t, err)

	keysOther := make([][]byte, len(keys))
	for i := range keysOther {
		keysOther[i] = bytes.Clone(keys[i])
		keysOther[i][0]++
	}

	r.SetPassedStorageNodes(keysOther)
	require.Equal(t, keysOther, r.PassedStorageNodes())

	err = r.Unmarshal(b)
	require.NoError(t, err)
	require.Equal(t, keys, r.PassedStorageNodes())
}

func TestResult_FailedStorageNodes(t *testing.T) {
	var r audit.Result
	r.ForContainer(cidtest.ID()) // just to satisfy Unmarshal
	var msg apiaudit.DataAuditResult
	keys := [][]byte{
		[]byte("any_key1"),
		[]byte("any_key2"),
	}

	require.Zero(t, r.FailedStorageNodes())

	b := r.Marshal()
	err := r.Unmarshal(b)
	require.NoError(t, err)
	require.Zero(t, r.FailedStorageNodes())

	r.SetFailedStorageNodes(keys)
	require.Equal(t, keys, r.FailedStorageNodes())

	b = r.Marshal()
	err = proto.Unmarshal(b, &msg)
	require.NoError(t, err)

	b, err = proto.Marshal(&msg)
	require.NoError(t, err)

	keysOther := make([][]byte, len(keys))
	for i := range keysOther {
		keysOther[i] = bytes.Clone(keys[i])
		keysOther[i][0]++
	}

	r.SetFailedStorageNodes(keysOther)
	require.Equal(t, keysOther, r.FailedStorageNodes())

	err = r.Unmarshal(b)
	require.NoError(t, err)
	require.Equal(t, keys, r.FailedStorageNodes())
}
