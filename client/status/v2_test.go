package apistatus_test

import (
	"errors"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/status"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/stretchr/testify/require"
)

func TestErrorToV2(t *testing.T) {
	for _, testCase := range []struct {
		e       error
		code    uint32
		msg     string
		details []*status.Status_Detail
	}{
		{nil, 0, "", nil},
		{errors.New("some error"), 1024, "some error", nil},
		{apistatus.ErrServerInternal, 1024, "", nil},
		{apistatus.NewInternalServerError(errors.New("some reason")), 1024, "some reason", nil},
		{apistatus.ErrWrongNetMagic, 1025, "", nil},
		{apistatus.NewWrongNetMagicError(6091320843), 1025, "",
			[]*status.Status_Detail{{Id: 0, Value: []byte{0, 0, 0, 1, 107, 18, 46, 11}}}},
		{apistatus.ErrSignatureVerification, 1026, "", nil},
		{apistatus.NewSignatureVerificationFailure(errors.New("some reason")), 1026, "some reason", nil},
		{apistatus.ErrNodeUnderMaintenance, 1027, "", nil},
		{apistatus.ErrContainerNotFound, 3072, "", nil},
		{apistatus.NewContainerNotFoundError(errors.New("some reason")), 3072, "some reason", nil},
		{apistatus.ErrEACLNotFound, 3073, "", nil},
		{apistatus.NewEACLNotFoundError(errors.New("some reason")), 3073, "some reason", nil},
		{apistatus.ErrObjectLocked, 2050, "", nil},
		{apistatus.ErrLockIrregularObject, 2051, "", nil},
		{apistatus.ErrObjectAccessDenied, 2048, "", nil},
		{apistatus.NewObjectAccessDeniedError("some reason"), 2048, "",
			[]*status.Status_Detail{{Id: 0, Value: []byte("some reason")}}},
		{apistatus.ErrObjectNotFound, 2049, "", nil},
		{apistatus.NewObjectNotFoundError(errors.New("some reason")), 2049, "some reason", nil},
		{apistatus.ErrObjectAlreadyRemoved, 2052, "", nil},
		{apistatus.ErrObjectOutOfRange, 2053, "", nil},
		{apistatus.ErrSessionTokenNotFound, 4096, "", nil},
		{apistatus.ErrSessionTokenExpired, 4097, "", nil},
	} {
		st := apistatus.ErrorToV2(testCase.e)
		require.Equal(t, testCase.code, st.GetCode(), testCase)
		require.Equal(t, testCase.msg, st.GetMessage(), testCase)
		require.Equal(t, testCase.details, st.GetDetails(), testCase)
	}
}

func TestErrorFromV2(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		t.Run("unrecognized", func(t *testing.T) {
			st := &status.Status{
				Code: 92035,
				Details: []*status.Status_Detail{
					{Id: 45982, Value: []byte("detail_1")},
					{Id: 3464363, Value: []byte("detail_2")},
				},
			}
			e, err := apistatus.ErrorFromV2(st)
			require.NoError(t, err)
			require.ErrorIs(t, e, apistatus.Error)
			require.Equal(t, st, e.ErrorToV2())
			require.EqualError(t, e, "status: code = 92035 (unknown, details count = 2)")

			st.Message = "any message"
			e, err = apistatus.ErrorFromV2(st)
			require.NoError(t, err)
			require.Equal(t, st, e.ErrorToV2())
			require.EqualError(t, e, "status: code = 92035 (unknown, details count = 2) message = any message")
		})
		for _, testCase := range []struct {
			st  *status.Status
			err error
		}{
			{st: nil, err: nil},
			{st: new(status.Status), err: nil},
			{st: &status.Status{Code: 1024}, err: apistatus.ErrServerInternal},
			{st: &status.Status{Code: 1024, Message: "some reason"},
				err: apistatus.NewInternalServerError(errors.New("some reason"))},
			{st: &status.Status{Code: 1025}, err: apistatus.ErrWrongNetMagic},
			{st: &status.Status{Code: 1025, Details: []*status.Status_Detail{{Id: 0, Value: []byte{0, 0, 0, 1, 107, 18, 46, 11}}}},
				err: apistatus.NewWrongNetMagicError(6091320843)},
			{st: &status.Status{Code: 1026}, err: apistatus.ErrSignatureVerification},
			{st: &status.Status{Code: 1026, Message: "some reason"},
				err: apistatus.NewSignatureVerificationFailure(errors.New("some reason"))},
			{st: &status.Status{Code: 1027}, err: apistatus.ErrNodeUnderMaintenance},
			{st: &status.Status{Code: 3072}, err: apistatus.ErrContainerNotFound},
			{st: &status.Status{Code: 3072, Message: "some reason"},
				err: apistatus.NewContainerNotFoundError(errors.New("some reason"))},
			{st: &status.Status{Code: 3073}, err: apistatus.ErrEACLNotFound},
			{st: &status.Status{Code: 3073, Message: "some reason"},
				err: apistatus.NewEACLNotFoundError(errors.New("some reason"))},
			{st: &status.Status{Code: 2050}, err: apistatus.ErrObjectLocked},
			{st: &status.Status{Code: 2051}, err: apistatus.ErrLockIrregularObject},
			{st: &status.Status{Code: 2048}, err: apistatus.ErrObjectAccessDenied},
			{st: &status.Status{Code: 2048, Details: []*status.Status_Detail{{Id: 0, Value: []byte("some reason")}}},
				err: apistatus.NewObjectAccessDeniedError("some reason")},
			{st: &status.Status{Code: 2049}, err: apistatus.ErrObjectNotFound},
			{st: &status.Status{Code: 2049, Message: "some reason"},
				err: apistatus.NewObjectNotFoundError(errors.New("some reason"))},
			{st: &status.Status{Code: 2052}, err: apistatus.ErrObjectAlreadyRemoved},
			{st: &status.Status{Code: 2053}, err: apistatus.ErrObjectOutOfRange},
			{st: &status.Status{Code: 4096}, err: apistatus.ErrSessionTokenNotFound},
			{st: &status.Status{Code: 4097}, err: apistatus.ErrSessionTokenExpired},
		} {
			e, err := apistatus.ErrorFromV2(testCase.st)
			require.NoError(t, err)
			require.Equal(t, testCase.err, e, testCase)
		}
	})
	t.Run("invalid", func(t *testing.T) {
		for _, testCase := range []struct {
			e  string
			st *status.Status
		}{
			{e: "invalid internal server error status: details attached but not supported",
				st: &status.Status{Code: 1024, Details: make([]*status.Status_Detail, 1)},
			},
			{e: "invalid wrong network magic status: too many details (2)",
				st: &status.Status{Code: 1025, Details: make([]*status.Status_Detail, 2)},
			},
			{e: "invalid wrong network magic status: unsupported detail ID=42",
				st: &status.Status{Code: 1025, Details: []*status.Status_Detail{{Id: 42}}},
			},
			{e: "invalid wrong network magic status: invalid correct value detail: invalid length 7",
				st: &status.Status{Code: 1025, Details: []*status.Status_Detail{{Id: 0, Value: make([]byte, 7)}}},
			},
			{e: "invalid wrong network magic status: invalid correct value detail: invalid length 9",
				st: &status.Status{Code: 1025, Details: []*status.Status_Detail{{Id: 0, Value: make([]byte, 9)}}},
			},
			{e: "invalid signature verification failure status: details attached but not supported",
				st: &status.Status{Code: 1026, Details: make([]*status.Status_Detail, 1)},
			},
			{e: "invalid node maintenance status: details attached but not supported",
				st: &status.Status{Code: 1027, Details: make([]*status.Status_Detail, 1)},
			},
			{e: "invalid missing container status: details attached but not supported",
				st: &status.Status{Code: 3072, Details: make([]*status.Status_Detail, 1)},
			},
			{e: "invalid missing eACL status: details attached but not supported",
				st: &status.Status{Code: 3073, Details: make([]*status.Status_Detail, 1)},
			},
			{e: "invalid locked object status: details attached but not supported",
				st: &status.Status{Code: 2050, Details: make([]*status.Status_Detail, 1)},
			},
			{e: "invalid locking irregular object status: details attached but not supported",
				st: &status.Status{Code: 2051, Details: make([]*status.Status_Detail, 1)},
			},
			{e: "invalid object access denial status: too many details (2)",
				st: &status.Status{Code: 2048, Details: make([]*status.Status_Detail, 2)},
			},
			{e: "invalid object access denial status: unsupported detail ID=42",
				st: &status.Status{Code: 2048, Details: []*status.Status_Detail{{Id: 42}}},
			},
			{e: "invalid object access denial status: invalid reason detail: invalid UTF-8 string",
				st: &status.Status{Code: 2048, Details: []*status.Status_Detail{{Id: 0, Value: []byte{66, 250, 67}}}},
			},
			{e: "invalid missing object status: details attached but not supported",
				st: &status.Status{Code: 2049, Details: make([]*status.Status_Detail, 1)},
			},
			{e: "invalid already removed object status: details attached but not supported",
				st: &status.Status{Code: 2052, Details: make([]*status.Status_Detail, 1)},
			},
			{e: "invalid out-of-range status: details attached but not supported",
				st: &status.Status{Code: 2053, Details: make([]*status.Status_Detail, 1)},
			},
			{e: "invalid missing session token status: details attached but not supported",
				st: &status.Status{Code: 4096, Details: make([]*status.Status_Detail, 1)},
			},
			{e: "invalid expired session token status: details attached but not supported",
				st: &status.Status{Code: 4097, Details: make([]*status.Status_Detail, 1)},
			},
		} {
			_, err := apistatus.ErrorFromV2(testCase.st)
			require.EqualError(t, err, testCase.e, testCase)
		}
	})
}
