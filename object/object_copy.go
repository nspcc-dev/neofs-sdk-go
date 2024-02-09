package object

import (
	"github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
)

func copyByteSlice(sl []byte) []byte {
	if sl == nil {
		return nil
	}

	bts := make([]byte, len(sl))
	copy(bts, sl)
	return bts
}

func copyObjectID(id *refs.ObjectID) *refs.ObjectID {
	if id == nil {
		return nil
	}

	var newID refs.ObjectID
	newID.SetValue(copyByteSlice(id.GetValue()))

	return &newID
}

func copySignature(sig *refs.Signature) *refs.Signature {
	if sig == nil {
		return nil
	}

	var newSig refs.Signature
	newSig.SetScheme(sig.GetScheme())
	newSig.SetKey(copyByteSlice(sig.GetKey()))
	newSig.SetSign(copyByteSlice(sig.GetSign()))

	return &newSig
}

func copySession(session *v2session.Token) *v2session.Token {
	if session == nil {
		return nil
	}

	var newSession v2session.Token
	if body := session.GetBody(); body != nil {
		var newBody v2session.TokenBody
		newBody.SetID(copyByteSlice(body.GetID()))

		if ownerID := body.GetOwnerID(); ownerID != nil {
			var newOwnerID refs.OwnerID
			newOwnerID.SetValue(copyByteSlice(ownerID.GetValue()))

			newBody.SetOwnerID(&newOwnerID)
		} else {
			newBody.SetOwnerID(nil)
		}

		if lifetime := body.GetLifetime(); lifetime != nil {
			newLifetime := *lifetime
			newBody.SetLifetime(&newLifetime)
		} else {
			newBody.SetLifetime(nil)
		}

		newBody.SetSessionKey(copyByteSlice(body.GetSessionKey()))

		// it is an interface. Both implementations do nothing inside implemented functions.
		newBody.SetContext(body.GetContext())

		newSession.SetBody(&newBody)
	} else {
		newSession.SetBody(nil)
	}

	newSession.SetSignature(copySignature(session.GetSignature()))

	return &newSession
}

func copySplitHeader(spl *object.SplitHeader) *object.SplitHeader {
	if spl == nil {
		return nil
	}

	var newSpl object.SplitHeader

	newSpl.SetParent(copyObjectID(spl.GetParent()))
	newSpl.SetPrevious(copyObjectID(spl.GetPrevious()))
	newSpl.SetFirst(copyObjectID(spl.GetFirst()))
	newSpl.SetParentSignature(copySignature(spl.GetParentSignature()))
	newSpl.SetParentHeader(copyHeader(spl.GetParentHeader()))

	if children := spl.GetChildren(); children != nil {
		newChildren := make([]refs.ObjectID, len(children))
		copy(newChildren, children)

		newSpl.SetChildren(newChildren)
	} else {
		newSpl.SetChildren(nil)
	}

	newSpl.SetSplitID(copyByteSlice(spl.GetSplitID()))

	return &newSpl
}

func copyHeader(header *object.Header) *object.Header {
	if header == nil {
		return nil
	}

	var newHeader object.Header

	newHeader.SetCreationEpoch(header.GetCreationEpoch())
	newHeader.SetPayloadLength(header.GetPayloadLength())
	newHeader.SetObjectType(header.GetObjectType())

	if ver := header.GetVersion(); ver != nil {
		newVer := *ver
		newHeader.SetVersion(&newVer)
	} else {
		newHeader.SetVersion(nil)
	}

	if containerID := header.GetContainerID(); containerID != nil {
		var newContainerID refs.ContainerID
		newContainerID.SetValue(copyByteSlice(containerID.GetValue()))

		newHeader.SetContainerID(&newContainerID)
	} else {
		newHeader.SetContainerID(nil)
	}

	if ownerID := header.GetOwnerID(); ownerID != nil {
		var newOwnerID refs.OwnerID
		newOwnerID.SetValue(copyByteSlice(ownerID.GetValue()))

		newHeader.SetOwnerID(&newOwnerID)
	} else {
		newHeader.SetOwnerID(nil)
	}

	if payloadHash := header.GetPayloadHash(); payloadHash != nil {
		var newPayloadHash refs.Checksum
		newPayloadHash.SetType(payloadHash.GetType())
		newPayloadHash.SetSum(copyByteSlice(payloadHash.GetSum()))

		newHeader.SetPayloadHash(&newPayloadHash)
	} else {
		newHeader.SetPayloadHash(nil)
	}

	if homoHash := header.GetHomomorphicHash(); homoHash != nil {
		var newHomoHash refs.Checksum
		newHomoHash.SetType(homoHash.GetType())
		newHomoHash.SetSum(copyByteSlice(homoHash.GetSum()))

		newHeader.SetHomomorphicHash(&newHomoHash)
	} else {
		newHeader.SetHomomorphicHash(nil)
	}

	newHeader.SetSessionToken(copySession(header.GetSessionToken()))

	if attrs := header.GetAttributes(); attrs != nil {
		newAttributes := make([]object.Attribute, len(attrs))
		copy(newAttributes, attrs)

		newHeader.SetAttributes(newAttributes)
	} else {
		newHeader.SetAttributes(nil)
	}

	newHeader.SetSplit(copySplitHeader(header.GetSplit()))

	return &newHeader
}
