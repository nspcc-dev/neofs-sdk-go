package object

// SplitInfoError is a special error that means that the original object is a large one (split into a number of smaller objects).
type SplitInfoError struct {
	si *SplitInfo
}

const splitInfoErrorMsg = "object not found, split info has been provided"

// Error implements the error interface.
func (s *SplitInfoError) Error() string {
	return splitInfoErrorMsg
}

// SplitInfo returns [SplitInfo] data.
func (s *SplitInfoError) SplitInfo() *SplitInfo {
	return s.si
}

// NewSplitInfoError is a constructor for [SplitInfoError].
func NewSplitInfoError(v *SplitInfo) *SplitInfoError {
	return &SplitInfoError{si: v}
}
