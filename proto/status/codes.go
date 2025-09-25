package status

// All supported status codes.
const (
	OK                        = 0
	IncompleteSuccess         = 1
	InternalServerError       = 1024
	WrongNetMagic             = 1025
	SignatureVerificationFail = 1026
	NodeUnderMaintenance      = 1027
	ObjectAccessDenied        = 2048
	ObjectNotFound            = 2049
	ObjectLocked              = 2050
	LockIrregularObject       = 2051
	ObjectAlreadyRemoved      = 2052
	OutOfRange                = 2053
	QuotaExceeded             = 2054
	ContainerNotFound         = 3072
	EACLNotFound              = 3073
	SessionTokenNotFound      = 4096
	SessionTokenExpired       = 4097
)

// All supported status details.
const (
	DetailCorrectNetMagic          = 0
	DetailObjectAccessDenialReason = 0
)
