package version

import (
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
)

// Version represents v2-compatible version.
type Version refs.Version

const sdkMjr, sdkMnr = 2, 11

// NewFromV2 wraps v2 Version message to Version.
//
// Nil refs.Version converts to nil.
func NewFromV2(v *refs.Version) *Version {
	return (*Version)(v)
}

// New creates and initializes blank Version.
//
// Defaults:
//  - major: 0;
//  - minor: 0.
func New() *Version {
	return NewFromV2(new(refs.Version))
}

// Current returns Version instance that
// initialized to current SDK revision number.
func Current() *Version {
	v := New()
	v.SetMajor(sdkMjr)
	v.SetMinor(sdkMnr)

	return v
}

// Major returns major number of the revision.
func (v *Version) Major() uint32 {
	return (*refs.Version)(v).GetMajor()
}

// SetMajor sets major number of the revision.
func (v *Version) SetMajor(val uint32) {
	(*refs.Version)(v).SetMajor(val)
}

// Minor returns minor number of the revision.
func (v *Version) Minor() uint32 {
	return (*refs.Version)(v).GetMinor()
}

// SetMinor sets minor number of the revision.
func (v *Version) SetMinor(val uint32) {
	(*refs.Version)(v).SetMinor(val)
}

// ToV2 converts Version to v2 Version message.
//
// Nil Version converts to nil.
func (v *Version) ToV2() *refs.Version {
	return (*refs.Version)(v)
}

func (v *Version) String() string {
	return fmt.Sprintf("v%d.%d", v.Major(), v.Minor())
}

// Marshal marshals Version into a protobuf binary form.
func (v *Version) Marshal() ([]byte, error) {
	return (*refs.Version)(v).StableMarshal(nil), nil
}

// Unmarshal unmarshals protobuf binary representation of Version.
func (v *Version) Unmarshal(data []byte) error {
	return (*refs.Version)(v).Unmarshal(data)
}

// MarshalJSON encodes Version to protobuf JSON format.
func (v *Version) MarshalJSON() ([]byte, error) {
	return (*refs.Version)(v).MarshalJSON()
}

// UnmarshalJSON decodes Version from protobuf JSON format.
func (v *Version) UnmarshalJSON(data []byte) error {
	return (*refs.Version)(v).UnmarshalJSON(data)
}
