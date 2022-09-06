package object

const (
	// AttributeName is an attribute key that is commonly used to denote
	// human-friendly name.
	AttributeName = "Name"

	// AttributeFileName is an attribute key that is commonly used to denote
	// file name to be associated with the object on saving.
	AttributeFileName = "FileName"

	// AttributeFilePath is an attribute key that is commonly used to denote
	// full path to be associated with the object on saving. Should start with a
	// '/' and use '/' as a delimiting symbol. Trailing '/' should be
	// interpreted as a virtual directory marker. If an object has conflicting
	// FilePath and FileName, FilePath should have higher priority, because it
	// is used to construct the directory tree. FilePath with trailing '/' and
	// non-empty FileName attribute should not be used together.
	AttributeFilePath = "FilePath"

	// AttributeTimestamp is an attribute key that is commonly used to denote
	// user-defined local time of object creation in Unix Timestamp format.
	AttributeTimestamp = "Timestamp"

	// AttributeContentType is an attribute key that is commonly used to denote
	// MIME Content Type of object's payload.
	AttributeContentType = "Content-Type"
)
