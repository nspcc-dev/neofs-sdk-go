package accounting_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/proto/accounting"
	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
)

// returns random accounting.Decimal with all non-zero fields.
func randDecimal() *accounting.Decimal {
	return &accounting.Decimal{Value: prototest.RandInt64(), Precision: prototest.RandUint32()}
}

func TestDecimal_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*accounting.Decimal{
		{Value: prototest.RandInt64()},
		{Precision: prototest.RandUint32()},
		randDecimal(),
	})
}

func TestBalanceRequest_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*accounting.BalanceRequest_Body{
		{OwnerId: new(refs.OwnerID)},
		{OwnerId: &refs.OwnerID{Value: []byte{}}},
		{OwnerId: prototest.RandOwnerID()},
	})
}

func TestBalanceResponse_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*accounting.BalanceResponse_Body{
		{Balance: new(accounting.Decimal)},
		{Balance: randDecimal()},
	})
}
