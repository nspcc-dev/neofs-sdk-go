package acl_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/proto/acl"
	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
)

func TestEACLRecord_Filter_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*acl.EACLRecord_Filter{
		prototest.RandEACLFilter(),
	})
}

func TestEACLRecord_Target_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*acl.EACLRecord_Target{
		prototest.RandEACLTarget(),
	})
}

func TestEACLRecord_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*acl.EACLRecord{
		prototest.RandEACLRecord(),
	})
}

func TestEACLTable_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*acl.EACLTable{
		prototest.RandEACL(),
	})
}

func TestBearerToken_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*acl.BearerToken_Body{
		prototest.RandBearerTokenBody(),
	})
}

func TestBearerToken_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*acl.BearerToken{
		prototest.RandBearerToken(),
	})
}
