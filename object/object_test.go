package object_test

import (
	"strconv"
	"testing"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func TestInitCreation(t *testing.T) {
	var o object.Object
	cnr := cidtest.ID()
	own := usertest.ID()

	o.InitCreation(object.RequiredFields{
		Container: cnr,
		Owner:     own,
	})

	cID, set := o.ContainerID()
	require.True(t, set)
	require.Equal(t, cnr, cID)
	require.Equal(t, &own, o.OwnerID())
}

func TestObject_UserAttributes(t *testing.T) {
	var obj object.Object
	var attrs []object.Attribute
	mSys := make(map[string]string)
	mUsr := make(map[string]string)

	for i := 0; i < 10; i++ {
		si := strconv.Itoa(i)

		keyUsr := "key" + si
		valUsr := "val" + si
		keySys := "__NEOFS__" + si
		valSys := "sys-val" + si

		mUsr[keyUsr] = valUsr
		mSys[keySys] = valSys

		var aUsr object.Attribute
		aUsr.SetKey(keyUsr)
		aUsr.SetValue(valUsr)

		var aSys object.Attribute
		aSys.SetKey(keySys)
		aSys.SetValue(valSys)

		attrs = append(attrs, aSys, aUsr)
	}

	obj.SetAttributes(attrs...)

	for _, a := range obj.UserAttributes() {
		key, val := a.Key(), a.Value()
		_, isSys := mSys[key]
		require.False(t, isSys, key)
		require.Equal(t, mUsr[key], val, key)
		delete(mUsr, key)
	}

	require.Empty(t, mUsr)
}
