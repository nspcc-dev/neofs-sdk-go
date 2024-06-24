package session

import (
	"bytes"
	"testing"

	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func TestCopyTo(t *testing.T) {
	usr, _ := usertest.TwoUsers()

	var cnr Container
	var obj Object
	cnr.SetAuthKey(usr.Public())
	obj.SetAuthKey(usr.Public())
	obj.LimitByObjects(oidtest.NIDs(3))

	objsCp := make([]oid.ID, len(obj.objs))
	copy(objsCp, obj.objs)
	authKeyCp := bytes.Clone(obj.authKey)

	cnrShallow := cnr
	objShallow := obj
	var cnrDeep Container
	cnr.CopyTo(&cnrDeep)
	var objDeep Object
	obj.CopyTo(&objDeep)
	require.Equal(t, cnr, cnrShallow)
	require.Equal(t, cnr, cnrDeep)
	require.Equal(t, obj, objShallow)
	require.Equal(t, obj, objDeep)

	cnr.authKey[0]++
	obj.authKey[0]++
	require.Equal(t, cnr.authKey, cnrShallow.authKey)
	require.Equal(t, obj.authKey, objShallow.authKey)
	require.Equal(t, authKeyCp, cnrDeep.authKey)
	require.Equal(t, authKeyCp, objDeep.authKey)

	obj.objs[1][0]++
	require.Equal(t, obj.objs, objShallow.objs)
	require.Equal(t, objsCp, objDeep.objs)
}
