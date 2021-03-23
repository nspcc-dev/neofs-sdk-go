package object

// SplitInfoError is an error returned
// on split object encounter.
type SplitInfoError struct {
	si *SplitInfo
}

const splitInfoErrorMsg = "object not found, split info has been provided"

func (s *SplitInfoError) Error() string {
	return splitInfoErrorMsg
}

// SplitInfo returns object SplitInfo.
func (s *SplitInfoError) SplitInfo() *SplitInfo {
	return s.si
}

// NewSplitInfoError creates new SplitInfoError
// with provided SplitInfo.
func NewSplitInfoError(v *SplitInfo) error {
	return &SplitInfoError{si: v}
}
