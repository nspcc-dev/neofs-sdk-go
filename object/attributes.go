package object

import (
	"fmt"
	"strconv"
	"time"
)

// Various attributes popular in applications.
const (
	attributeName        = "Name"
	attributeFileName    = "FileName"
	attributeFilePath    = "FilePath"
	attributeTimestamp   = "Timestamp"
	attributeContentType = "Content-Type"
)

// SetName associates given name with the object by setting its 'Name'
// attribute. The name must not be empty and is expected to be human-readable.
// Use [GetName] to read and [FilterByName] to search. The property is treated
// by the system as a regular user attribute.
func SetName(obj objectOrHeaderPtr, name string) {
	obj.SetAttribute(attributeName, name)
}

// GetName returns object name set according to [SetName]. Zero return means
// unset name.
func GetName(obj objectOrHeader) string {
	return obj.Attribute(attributeName)
}

// SetFileName associates given file name with the object by setting its
// 'FileName' attribute. The file name must not be empty. Use [GetFileName] to
// read and [FilterByFileName] to search. The property is treated by the system as a
// regular user attribute.
func SetFileName(obj objectOrHeaderPtr, file string) {
	obj.SetAttribute(attributeFileName, file)
}

// GetFileName returns associated file name set according to [SetFileName]. Zero
// return means unset file name.
func GetFileName(obj objectOrHeader) string {
	return obj.Attribute(attributeFileName)
}

// SetFilePath associates given filesystem path with the object by settings its
// 'FilePath' attribute. The path must not be empty. Use [GetFilePath] to read
// and [FilterByFilePath] to search. The property is treated by the system as a
// regular user attribute.
//
// The file path should start with a '/' and use '/' as a delimiting symbol.
// Trailing '/' should be interpreted as a virtual directory marker. If an
// object has conflicting file path and name (see [SetFileName]), the first one
// should have higher priority because it is used to construct the directory
// tree. The file path with trailing '/' and non-empty name should not be used
// together. Again, these statements are purely advisory and are not verified by
// the system.
func SetFilePath(obj objectOrHeaderPtr, filePath string) {
	obj.SetAttribute(attributeFilePath, filePath)
}

// GetFilePath returns associated file name set according to [SetFilePath]. Zero
// return means unset file name.
func GetFilePath(obj objectOrHeader) string {
	return obj.Attribute(attributeFilePath)
}

// SetCreationTime stamps object's creation time in Unix Timestamp format by
// setting its 'Timestamp' attribute. Use [GetCreationTime] to read and
// [FilterByCreationTime] to search. The property is treated by the system as a
// regular user attribute.
func SetCreationTime(obj objectOrHeaderPtr, t time.Time) {
	obj.SetAttribute(attributeTimestamp, strconv.FormatInt(t.Unix(), 10))
}

// GetCreationTime returns object's creation time set according to
// [SetCreationTime]. Zero return (in seconds) means unset timestamp.
func GetCreationTime(obj objectOrHeader) time.Time {
	var sec int64
	if s := obj.Attribute(attributeTimestamp); s != "" {
		var err error
		sec, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("parse timestamp attribute: %v", err))
		}
	}
	return time.Unix(sec, 0)
}

// SetContentType specifies MIME content type of payload of the object by
// setting its 'Content-Type' attribute. The type must not be empty. Use
// [GetContentType] to read and [FilterByContentType] to search. The property is
// treated by the system as a regular user attribute.
func SetContentType(obj objectOrHeaderPtr, contentType string) {
	obj.SetAttribute(attributeContentType, contentType)
}

// GetContentType returns content type of the object payload set according to
// [SetContentType]. Zero return means unset content type.
func GetContentType(obj objectOrHeader) string {
	return obj.Attribute(attributeContentType)
}
