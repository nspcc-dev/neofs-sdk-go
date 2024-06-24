package accounting

import (
	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
)

const (
	_ = iota
	fieldDecimalValue
	fieldDecimalPrecision
)

func (x *Decimal) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldDecimalValue, x.Value) +
			proto.SizeVarint(fieldDecimalPrecision, x.Precision)
	}
	return sz
}

func (x *Decimal) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldDecimalValue, x.Value)
		proto.MarshalVarint(b[off:], fieldDecimalPrecision, x.Precision)
	}
}

const (
	_ = iota
	fieldBalanceReqOwner
)

func (x *BalanceRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldBalanceReqOwner, x.OwnerId)
	}
	return sz
}

func (x *BalanceRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalNested(b, fieldBalanceReqOwner, x.OwnerId)
	}
}

const (
	_ = iota
	fieldBalanceRespBalance
)

func (x *BalanceResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldBalanceRespBalance, x.Balance)
	}
	return sz
}

func (x *BalanceResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalNested(b, fieldBalanceRespBalance, x.Balance)
	}
}
