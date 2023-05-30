package poolv2

import (
	"fmt"
	"io"

	netmapApiGo "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

type CreateContainerBuilder struct {
	basicACL acl.Basic
	owner    user.ID
	attr     []Attribute

	// could be placed to another builder
	placementPolicyV2 netmapApiGo.PlacementPolicy
	filters           []netmapApiGo.Filter
	selectors         []netmapApiGo.Selector
	replicas          []netmapApiGo.Replica
	backupFactor      uint32
}

func NewCreateContainerBuilder(owner user.ID) *CreateContainerBuilder {
	return &CreateContainerBuilder{
		basicACL:     acl.PublicRW,
		owner:        owner,
		backupFactor: 1,
	}
}

func (b *CreateContainerBuilder) WithAcl(basicACL acl.Basic) *CreateContainerBuilder {
	b.basicACL = basicACL
	return b
}

func (b *CreateContainerBuilder) WithOwner(owner user.ID) *CreateContainerBuilder {
	b.owner = owner
	return b
}

func (b *CreateContainerBuilder) WithAttributes(attr []Attribute) *CreateContainerBuilder {
	b.attr = attr
	return b
}

func (b *CreateContainerBuilder) WithAttribute(attr Attribute) *CreateContainerBuilder {
	b.attr = append(b.attr, attr)
	return b
}

func (b *CreateContainerBuilder) WithPPFilter(name, value, key string, op netmapApiGo.Operation, filters []netmapApiGo.Filter) *CreateContainerBuilder {
	f := netmapApiGo.Filter{}
	f.SetName(name)
	f.SetValue(value)
	f.SetOp(op)
	f.SetKey(key)
	f.SetFilters(filters)

	b.filters = append(b.filters, f)
	return b
}

func (b *CreateContainerBuilder) WithPPBackupFactor(f uint32) *CreateContainerBuilder {
	b.backupFactor = f
	return b
}

func (b *CreateContainerBuilder) placementPolicy() (netmap.PlacementPolicy, error) {
	var placementPolicyV2 netmapApiGo.PlacementPolicy

	placementPolicyV2.SetFilters(b.filters)
	placementPolicyV2.SetSelectors(b.selectors)
	placementPolicyV2.SetReplicas(b.replicas)

	var placementPolicy netmap.PlacementPolicy
	if err := placementPolicy.ReadFromV2(placementPolicyV2); err != nil {
		return placementPolicy, fmt.Errorf("invalid argument: %w", err)
	}

	placementPolicy.SetContainerBackupFactor(b.backupFactor)
	return placementPolicy, nil
}

type CreateObjectBuilder struct {
	payload     io.Reader
	attr        []Attribute
	idHandler   IDHandler
	containerID cid.ID
}

func NewCreateObjectBuilder(containerID cid.ID, payload io.Reader) *CreateObjectBuilder {
	return &CreateObjectBuilder{
		containerID: containerID,
		payload:     payload,
	}
}

func (b *CreateObjectBuilder) WithAttributes(attr []Attribute) *CreateObjectBuilder {
	b.attr = attr
	return b
}

func (b *CreateObjectBuilder) WithAttribute(attr Attribute) *CreateObjectBuilder {
	b.attr = append(b.attr, attr)
	return b
}

func (b *CreateObjectBuilder) WithPayload(payload io.Reader) *CreateObjectBuilder {
	b.payload = payload
	return b
}

func (b *CreateObjectBuilder) WithIDHandler(idHandler IDHandler) *CreateObjectBuilder {
	b.idHandler = idHandler
	return b
}

func (b *CreateObjectBuilder) InContainer(containerID cid.ID) *CreateObjectBuilder {
	b.containerID = containerID
	return b
}

type ReadObjectBuilder struct {
	containerID cid.ID
	objectID    oid.ID
	writer      io.Writer
}

func NewReadObjectBuilder(containerID cid.ID, objectID oid.ID, writer io.Writer) *ReadObjectBuilder {
	return &ReadObjectBuilder{
		containerID: containerID,
		objectID:    objectID,
		writer:      writer,
	}
}
