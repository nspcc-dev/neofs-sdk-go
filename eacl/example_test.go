package eacl_test

import (
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/object"
)

func randomPublicKey() neofscrypto.PublicKey {
	k, err := keys.NewPrivateKey()
	if err != nil {
		panic(fmt.Errorf("randomize private key: %v", err))
	}

	return neofsecdsa.Signer(k.PrivateKey).Public()
}

// eACL provides ability to determine target subjects of access rules, in
// particular, distribute/restrict access on an exclusive basis.
func ExampleTable_exclusiveRights() {
	friendPubKey := randomPublicKey()
	vacationPhotos := eacl.NewFilterObjectAttribute(object.AttributeFileName, eacl.MatchStringEqual, "vacation_photos.zip")

	_ = eacl.New([]eacl.Record{
		// allowing rule goes first because order matters here: friend is most likely eacl.RoleOthers
		eacl.NewRecord(eacl.ActionAllow, acl.OpObjectGet, eacl.NewTargetWithKey(friendPubKey), vacationPhotos),
		eacl.NewRecord(eacl.ActionDeny, acl.OpObjectGet, eacl.NewTargetWithRole(eacl.RoleOthers), vacationPhotos),
	})

	cvApplication := eacl.NewFilterObjectAttribute(object.AttributeName, eacl.MatchStringEqual, "CV")
	bannedApplicantPubKey := randomPublicKey()
	managerPubKeys := []neofscrypto.PublicKey{
		randomPublicKey(),
		randomPublicKey(),
	}

	_ = eacl.New([]eacl.Record{
		// allowing rule goes before because order matters here: managers are most likely have eacl.RoleOthers
		eacl.NewRecord(eacl.ActionAllow, acl.OpObjectGet, eacl.NewTargetWithKeys(managerPubKeys), cvApplication),
		eacl.NewRecord(eacl.ActionDeny, acl.OpObjectGet, eacl.NewTargetWithRole(eacl.RoleOthers), cvApplication),
		// next pair of rules may be swapped in order with the previous one because they target different operations
		eacl.NewRecord(eacl.ActionDeny, acl.OpObjectPut, eacl.NewTargetWithKey(bannedApplicantPubKey), cvApplication),
		eacl.NewRecord(eacl.ActionAllow, acl.OpObjectPut, eacl.NewTargetWithRole(eacl.RoleOthers), cvApplication),
	})
}

// eACL allows to target particular resource operations with resource to control
// user actions.
func ExampleTable_objectOps() {
	// forbid user with this key to modify any object
	observerPubKey := randomPublicKey()
	// allow user with this public key to do everything with all objects
	adminPubKey := randomPublicKey()

	observerOnly := eacl.NewTargetWithKey(observerPubKey)
	adminOnly := eacl.NewTargetWithKey(adminPubKey)
	allOthers := eacl.NewTargetWithRole(eacl.RoleOthers)

	_ = eacl.New([]eacl.Record{
		eacl.NewRecord(eacl.ActionDeny, acl.OpObjectPut, observerOnly),
		eacl.NewRecord(eacl.ActionDeny, acl.OpObjectDelete, observerOnly),
		eacl.NewRecord(eacl.ActionAllow, acl.OpObjectGet, observerOnly),
		eacl.NewRecord(eacl.ActionAllow, acl.OpObjectHead, observerOnly),
		eacl.NewRecord(eacl.ActionAllow, acl.OpObjectRange, observerOnly),
		eacl.NewRecord(eacl.ActionAllow, acl.OpObjectSearch, observerOnly),
		eacl.NewRecord(eacl.ActionAllow, acl.OpObjectHash, observerOnly),

		eacl.NewRecord(eacl.ActionAllow, acl.OpObjectPut, adminOnly),
		eacl.NewRecord(eacl.ActionAllow, acl.OpObjectDelete, adminOnly),
		eacl.NewRecord(eacl.ActionAllow, acl.OpObjectGet, adminOnly),
		eacl.NewRecord(eacl.ActionAllow, acl.OpObjectHead, adminOnly),
		eacl.NewRecord(eacl.ActionAllow, acl.OpObjectRange, adminOnly),
		eacl.NewRecord(eacl.ActionAllow, acl.OpObjectSearch, adminOnly),
		eacl.NewRecord(eacl.ActionAllow, acl.OpObjectHash, adminOnly),

		eacl.NewRecord(eacl.ActionDeny, acl.OpObjectPut, allOthers),
		eacl.NewRecord(eacl.ActionDeny, acl.OpObjectDelete, allOthers),
		eacl.NewRecord(eacl.ActionDeny, acl.OpObjectGet, allOthers),
		eacl.NewRecord(eacl.ActionDeny, acl.OpObjectHead, allOthers),
		eacl.NewRecord(eacl.ActionDeny, acl.OpObjectRange, allOthers),
		eacl.NewRecord(eacl.ActionDeny, acl.OpObjectSearch, allOthers),
		eacl.NewRecord(eacl.ActionDeny, acl.OpObjectHash, allOthers),
	})
}

// Target access-controlled resources may be selected through descriptors. Only
// matching data is subject to the specified rules.
func ExampleTable_selectiveResources() {
	debugArtifactsOnly := eacl.NewFilterObjectAttribute("MODE", eacl.MatchStringEqual, "DEBUG")

	_ = eacl.New([]eacl.Record{
		eacl.NewRecord(eacl.ActionDeny, acl.OpObjectPut, eacl.NewTargetWithRole(eacl.RoleOthers), debugArtifactsOnly),
		eacl.NewRecord(eacl.ActionAllow, acl.OpObjectPut, eacl.NewTargetWithRole(eacl.RoleOthers)),
	})

	// filters are imposed with AND operator
	_ = eacl.New([]eacl.Record{
		eacl.NewRecord(eacl.ActionDeny, acl.OpObjectDelete, eacl.NewTargetWithRole(eacl.RoleOthers),
			eacl.NewFilterObjectAttribute("Type", eacl.MatchStringEqual, "Document"),
			eacl.NewFilterObjectAttribute("Department", eacl.MatchStringEqual, "Archive")),
		eacl.NewRecord(eacl.ActionAllow, acl.OpObjectDelete, eacl.NewTargetWithRole(eacl.RoleOthers)),
	})
}
