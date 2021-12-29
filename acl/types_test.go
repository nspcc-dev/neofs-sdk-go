package acl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	for _, tc := range []struct {
		acl      string
		expected uint32
		err      bool
	}{
		{
			acl:      PublicBasicName,
			expected: PublicBasicRule,
		},
		{
			acl:      PrivateBasicName,
			expected: PrivateBasicRule,
		},
		{
			acl:      ReadOnlyBasicName,
			expected: ReadOnlyBasicRule,
		},
		{
			acl:      PublicAppendName,
			expected: PublicAppendRule,
		},
		{
			acl:      EACLPublicBasicName,
			expected: EACLPublicBasicRule,
		},
		{
			acl:      EACLPrivateBasicName,
			expected: EACLPrivateBasicRule,
		},
		{
			acl:      EACLReadOnlyBasicName,
			expected: EACLReadOnlyBasicRule,
		},
		{
			acl:      EACLPublicAppendName,
			expected: EACLPublicAppendRule,
		},
		{
			acl:      "0x1C8C8CCC",
			expected: 0x1C8C8CCC,
		},
		{
			acl:      "1C8C8CCC",
			expected: 0x1C8C8CCC,
		},
		{
			acl: "123456789",
			err: true,
		},
		{
			acl: "0x1C8C8CCG",
			err: true,
		},
	} {
		actual, err := ParseBasicACL(tc.acl)
		if tc.err {
			require.Error(t, err)
			continue
		}

		require.NoError(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestString(t *testing.T) {
	acl := BasicACL(0x1fbfbfff)
	require.Equal(t, "0x1fbfbfff", acl.String())

	acl2, err := ParseBasicACL(PrivateBasicName)
	require.NoError(t, err)
	require.Equal(t, "0x1c8c8ccc", BasicACL(acl2).String())
}
