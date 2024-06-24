package accounting_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/accounting"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestBalanceRequest_Body(t *testing.T) {
	v := &accounting.BalanceRequest_Body{
		OwnerId: &refs.OwnerID{
			Value: []byte("any_owner"),
		},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res accounting.BalanceRequest_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.OwnerId, res.OwnerId)
	// TODO: test field order. Pretty challenging in general, but can be simplified
	//  for NeoFS specifics (forbid group types, maps, etc.).
}

func TestBalanceResponse_Body(t *testing.T) {
	v := &accounting.BalanceResponse_Body{
		Balance: &accounting.Decimal{
			Value:     12,
			Precision: 34,
		},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res accounting.BalanceResponse_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Balance, res.Balance)
}
