package session

import (
	"bytes"
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// TokenV2CurrentVersion is the current TokenV2 version.
const TokenV2CurrentVersion = 1

// VerbV2 represents unified operations for both object and container services.
type VerbV2 int32

const (
	VerbV2Unspecified VerbV2 = iota
	VerbV2ObjectPut
	VerbV2ObjectGet
	VerbV2ObjectHead
	VerbV2ObjectSearch
	VerbV2ObjectDelete
	VerbV2ObjectRange
	VerbV2ObjectRangeHash
	VerbV2ContainerPut
	VerbV2ContainerDelete
	VerbV2ContainerSetEACL
)

// IsObjectVerb returns true if the verb is for object operations.
func (v VerbV2) IsObjectVerb() bool {
	return v >= VerbV2ObjectPut && v <= VerbV2ObjectRangeHash
}

// IsContainerVerb returns true if the verb is for container operations.
func (v VerbV2) IsContainerVerb() bool {
	return v >= VerbV2ContainerPut && v <= VerbV2ContainerSetEACL
}

// Lifetime represents token or delegation lifetime claims.
type Lifetime struct {
	iat uint64
	nbf uint64
	exp uint64
}

// NewLifetime creates a new Lifetime instance.
func NewLifetime(iat, nbf, exp uint64) Lifetime {
	return Lifetime{
		iat: iat,
		nbf: nbf,
		exp: exp,
	}
}

// Iat returns the issued at claim.
func (l Lifetime) Iat() uint64 {
	return l.iat
}

// Nbf returns the not before claim.
func (l Lifetime) Nbf() uint64 {
	return l.nbf
}

// Exp returns the expiration claim.
func (l Lifetime) Exp() uint64 {
	return l.exp
}

// SetIat sets the issued at claim.
func (l *Lifetime) SetIat(iat uint64) {
	l.iat = iat
}

// SetNbf sets the not before claim.
func (l *Lifetime) SetNbf(nbf uint64) {
	l.nbf = nbf
}

// SetExp sets the expiration claim.
func (l *Lifetime) SetExp(exp uint64) {
	l.exp = exp
}

// ValidAt checks if the lifetime is valid at the given timestamp.
func (l Lifetime) ValidAt(timestamp uint64) bool {
	return timestamp >= l.iat && timestamp <= l.exp && timestamp >= l.nbf
}

// Target represents an account identifier in session token V2.
// It can be either a direct OwnerID reference or an NNS name.
type Target struct {
	ownerID user.ID
	nnsName string
}

// NewTarget creates a new Target from user.ID.
func NewTarget(id user.ID) Target {
	return Target{ownerID: id}
}

// NewTargetFromNNS creates a new Target from NNS name.
func NewTargetFromNNS(nnsName string) Target {
	return Target{nnsName: nnsName}
}

// IsOwnerID returns true if target uses direct OwnerID reference.
func (t Target) IsOwnerID() bool {
	return !t.ownerID.IsZero()
}

// IsNNS returns true if target uses NNS name reference.
func (t Target) IsNNS() bool {
	return t.nnsName != ""
}

// OwnerID returns the owner ID if target uses direct reference.
func (t Target) OwnerID() user.ID {
	return t.ownerID
}

// NNSName returns the NNS name if target uses NNS reference.
func (t Target) NNSName() string {
	return t.nnsName
}

// Equals checks if two targets are equal.
func (t Target) Equals(other Target) bool {
	return t.ownerID == other.ownerID && t.nnsName == other.nnsName
}

// IsEmpty checks if the target is empty.
func (t Target) IsEmpty() bool {
	return t.ownerID.IsZero() && t.nnsName == ""
}

func (t Target) protoMessage() *protosession.Target {
	switch {
	case t.IsOwnerID():
		return &protosession.Target{
			Identifier: &protosession.Target_OwnerId{
				OwnerId: t.ownerID.ProtoMessage(),
			},
		}
	case t.IsNNS():
		return &protosession.Target{
			Identifier: &protosession.Target_NnsName{
				NnsName: t.nnsName,
			},
		}
	default:
		return nil
	}
}

func (t *Target) fromProtoMessage(m *protosession.Target) error {
	if m == nil {
		return errors.New("nil target")
	}

	switch id := m.Identifier.(type) {
	case *protosession.Target_OwnerId:
		if id.OwnerId == nil {
			return errors.New("nil owner ID in target")
		}
		return t.ownerID.FromProtoMessage(id.OwnerId)
	case *protosession.Target_NnsName:
		if id.NnsName == "" {
			return errors.New("empty NNS name in target")
		}
		t.nnsName = id.NnsName
		return nil
	default:
		return fmt.Errorf("unknown target identifier type: %T", id)
	}
}

// DelegationInfo represents a single delegation in the chain of trust.
type DelegationInfo struct {
	Lifetime
	issuer   Target
	subjects []Target
	verbs    []VerbV2

	sigSet bool
	sig    neofscrypto.Signature
}

// NewDelegationInfo creates a new delegation info.
// The issuer will be set automatically when Sign is called.
func NewDelegationInfo(subjects []Target, lifetime Lifetime, verbs []VerbV2) DelegationInfo {
	return DelegationInfo{
		subjects: slices.Clone(subjects),
		Lifetime: lifetime,
		verbs:    slices.Clone(verbs),
	}
}

// SetIssuer allows to set issuer before Sign call.
// Using this method is not required when Sign is used (issuer will be derived from the signer automatically).
// When using it please ensure that the delegation is signed with the same signer as the issuer passed here.
func (d *DelegationInfo) SetIssuer(issuer Target) {
	d.issuer = issuer
}

// Issuer returns the account that performed this delegation.
func (d DelegationInfo) Issuer() Target {
	return d.issuer
}

// Subjects returns the accounts that received the delegation.
func (d DelegationInfo) Subjects() []Target {
	return slices.Clone(d.subjects)
}

// Verbs returns the list of verbs authorized by this delegation.
func (d DelegationInfo) Verbs() []VerbV2 {
	return slices.Clone(d.verbs)
}

// Sign calculates and writes signature of the DelegationInfo data using signer.
func (d *DelegationInfo) Sign(signer user.Signer) error {
	d.issuer = NewTarget(signer.UserID())
	if d.issuer.IsEmpty() {
		return user.ErrZeroID
	}
	d.sigSet = true
	return d.sig.Calculate(signer, d.signedData())
}

// AttachSignature attaches given signature to the delegation info.
func (d *DelegationInfo) AttachSignature(sig neofscrypto.Signature) {
	d.sig, d.sigSet = sig, true
}

func (d DelegationInfo) signedData() []byte {
	m := &protosession.DelegationInfo{
		Issuer: d.issuer.protoMessage(),
		Verbs:  make([]protosession.Verb, len(d.verbs)),
	}

	if len(d.subjects) > 0 {
		m.Subjects = make([]*protosession.Target, len(d.subjects))
		for i := range d.subjects {
			m.Subjects[i] = d.subjects[i].protoMessage()
		}
	}

	if d.Iat() != 0 || d.Nbf() != 0 || d.Exp() != 0 {
		m.Lifetime = &protosession.TokenLifetime{
			Iat: d.Iat(),
			Nbf: d.Nbf(),
			Exp: d.Exp(),
		}
	}

	for i, v := range d.verbs {
		m.Verbs[i] = protosession.Verb(v)
	}
	return neofsproto.MarshalMessage(m)
}

// VerifySignature verifies the delegation signature.
func (d DelegationInfo) VerifySignature() bool {
	return d.sigSet && d.sig.Verify(d.signedData())
}

func (d DelegationInfo) protoMessage() *protosession.DelegationInfo {
	m := &protosession.DelegationInfo{
		Issuer: d.issuer.protoMessage(),
		Verbs:  make([]protosession.Verb, len(d.verbs)),
	}

	if len(d.subjects) > 0 {
		m.Subjects = make([]*protosession.Target, len(d.subjects))
		for i := range d.subjects {
			m.Subjects[i] = d.subjects[i].protoMessage()
		}
	}

	if d.Iat() != 0 || d.Nbf() != 0 || d.Exp() != 0 {
		m.Lifetime = &protosession.TokenLifetime{
			Iat: d.Iat(),
			Nbf: d.Nbf(),
			Exp: d.Exp(),
		}
	}

	for i, v := range d.verbs {
		m.Verbs[i] = protosession.Verb(v)
	}
	if d.sigSet {
		m.Signature = d.sig.ProtoMessage()
	}
	return m
}

func (d *DelegationInfo) fromProtoMessage(m *protosession.DelegationInfo) error {
	if m == nil {
		return errors.New("nil delegation info")
	}

	if err := d.issuer.fromProtoMessage(m.Issuer); err != nil {
		return fmt.Errorf("invalid issuer: %w", err)
	}

	subjects := m.GetSubjects()
	if len(subjects) > 0 {
		d.subjects = make([]Target, len(subjects))
		for i, subj := range subjects {
			if err := d.subjects[i].fromProtoMessage(subj); err != nil {
				return fmt.Errorf("invalid subject: %w", err)
			}
		}
	}

	lifetime := m.GetLifetime()
	if lifetime != nil {
		d.Lifetime = NewLifetime(
			lifetime.GetIat(),
			lifetime.GetNbf(),
			lifetime.GetExp(),
		)
	}

	d.verbs = make([]VerbV2, len(m.Verbs))
	for i, v := range m.Verbs {
		if v < 0 {
			return fmt.Errorf("negative verb %d", v)
		}
		d.verbs[i] = VerbV2(v)
	}

	if m.Signature != nil {
		d.sigSet = true
		if err := d.sig.FromProtoMessage(m.Signature); err != nil {
			return fmt.Errorf("invalid signature: %w", err)
		}
	}

	return nil
}

// ContextV2 represents a unified session context for both object and container operations.
type ContextV2 struct {
	container cid.ID
	objects   []oid.ID
	verbs     []VerbV2
}

// NewContextV2 creates a new unified session context.
func NewContextV2(container cid.ID, verbs []VerbV2) ContextV2 {
	return ContextV2{
		container: container,
		verbs:     slices.Clone(verbs),
	}
}

// Container returns the container ID for this context.
func (c ContextV2) Container() cid.ID {
	return c.container
}

// Objects returns the specific objects for this context.
func (c ContextV2) Objects() []oid.ID {
	return slices.Clone(c.objects)
}

// Verbs returns the authorized operations for this context.
func (c ContextV2) Verbs() []VerbV2 {
	return slices.Clone(c.verbs)
}

// SetObjects sets specific objects for this context.
func (c *ContextV2) SetObjects(objects []oid.ID) {
	c.objects = slices.Clone(objects)
}

func (c ContextV2) protoMessage() *protosession.SessionContextV2 {
	m := &protosession.SessionContextV2{
		Verbs: make([]protosession.Verb, len(c.verbs)),
	}

	if !c.container.IsZero() {
		m.Container = c.container.ProtoMessage()
	}

	if len(c.objects) > 0 {
		m.Objects = make([]*refs.ObjectID, len(c.objects))
		for i := range c.objects {
			m.Objects[i] = c.objects[i].ProtoMessage()
		}
	}

	for i, v := range c.verbs {
		m.Verbs[i] = protosession.Verb(v)
	}

	return m
}

func (c *ContextV2) fromProtoMessage(m *protosession.SessionContextV2) error {
	if m == nil {
		return errors.New("nil context")
	}

	if m.Container == nil && len(m.Objects) == 0 && len(m.Verbs) == 0 {
		return errors.New("empty context")
	}

	if m.Container != nil {
		if err := c.container.FromProtoMessage(m.Container); err != nil {
			return fmt.Errorf("invalid container: %w", err)
		}
	}

	if len(m.Objects) > 0 {
		c.objects = make([]oid.ID, len(m.Objects))
		for i, obj := range m.Objects {
			if obj == nil {
				return fmt.Errorf("nil object at index %d", i)
			}
			if err := c.objects[i].FromProtoMessage(obj); err != nil {
				return fmt.Errorf("invalid object at index %d: %w", i, err)
			}
		}
	}

	c.verbs = make([]VerbV2, len(m.Verbs))
	for i, v := range m.Verbs {
		if v < 0 {
			return fmt.Errorf("negative verb %d", v)
		}
		c.verbs[i] = VerbV2(v)
	}

	return nil
}

// TokenV2 represents NeoFS Session Token V2 with enhanced capabilities:
// - Multiple subjects (authorized users)
// - Multiple contexts (combined container and object operations)
// - Delegation chain tracking
// - NNS name resolution support
//
// TokenV2 is mutually compatible with [protosession.SessionTokenV2] message.
// See [TokenV2.FromProtoMessage] / [TokenV2.ProtoMessage] methods.
//
// Instances can be created using built-in var declaration.
type TokenV2 struct {
	Lifetime
	version  uint32
	idSet    bool
	id       uuid.UUID
	issuer   Target
	subjects []Target
	contexts []ContextV2

	delegationChain []DelegationInfo

	sigSet bool
	sig    neofscrypto.Signature
}

// CopyTo writes deep copy of the [TokenV2] to dst.
func (x TokenV2) CopyTo(dst *TokenV2) {
	dst.version = x.version
	dst.idSet = x.idSet
	dst.id = x.id
	dst.issuer = x.issuer
	dst.subjects = slices.Clone(x.subjects)
	dst.Lifetime = x.Lifetime
	dst.contexts = slices.Clone(x.contexts)
	dst.delegationChain = slices.Clone(x.delegationChain)
	dst.sigSet = x.sigSet
	if x.sigSet {
		dst.sig = neofscrypto.NewSignatureFromRawKey(x.sig.Scheme(), bytes.Clone(x.sig.PublicKeyBytes()), bytes.Clone(x.sig.Value()))
	}
}

// SetVersion sets the token version.
func (x *TokenV2) SetVersion(version uint32) {
	x.version = version
}

// Version returns the token version.
func (x TokenV2) Version() uint32 {
	return x.version
}

// SetID sets a unique identifier for the session token.
// ID format MUST be UUID version 4 (random).
func (x *TokenV2) SetID(id uuid.UUID) {
	x.id, x.idSet = id, true
}

// ID returns the session token identifier.
func (x TokenV2) ID() uuid.UUID {
	return x.id
}

// SetIssuer allows to set issuer before Sign call.
// Using this method is not required when Sign is used (issuer will be derived from the signer automatically).
// When using it please ensure that the token is signed with the same signer as the issuer passed here.
func (x *TokenV2) SetIssuer(issuer Target) {
	x.issuer = issuer
}

// Issuer returns the account that issued this token.
func (x TokenV2) Issuer() Target {
	return x.issuer
}

// SetSubjects sets the accounts authorized by this token.
func (x *TokenV2) SetSubjects(subjects []Target) {
	x.subjects = slices.Clone(subjects)
}

// Subjects returns the accounts authorized by this token.
func (x TokenV2) Subjects() []Target {
	return slices.Clone(x.subjects)
}

// AddSubject adds an authorized account to this token.
func (x *TokenV2) AddSubject(subject Target) {
	x.subjects = append(x.subjects, subject)
}

// SetContexts sets the session contexts.
func (x *TokenV2) SetContexts(contexts []ContextV2) {
	x.contexts = slices.Clone(contexts)
}

// Contexts returns the session contexts.
func (x TokenV2) Contexts() []ContextV2 {
	return slices.Clone(x.contexts)
}

// AddContext adds a session context.
func (x *TokenV2) AddContext(ctx ContextV2) {
	x.contexts = append(x.contexts, ctx)
}

// SetDelegationChain sets the full delegation chain.
func (x *TokenV2) SetDelegationChain(chain []DelegationInfo) {
	x.delegationChain = slices.Clone(chain)
}

// DelegationChain returns the full delegation chain.
func (x TokenV2) DelegationChain() []DelegationInfo {
	return slices.Clone(x.delegationChain)
}

// AddDelegation adds a delegation to the chain.
func (x *TokenV2) AddDelegation(delegation DelegationInfo) {
	x.delegationChain = append(x.delegationChain, delegation)
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// x from it.
func (x *TokenV2) FromProtoMessage(m *protosession.SessionTokenV2) error {
	return x.fromProtoMessage(m, true)
}

func (x *TokenV2) fromProtoMessage(m *protosession.SessionTokenV2, checkFieldPresence bool) error {
	body := m.GetBody()
	if checkFieldPresence && body == nil {
		return errors.New("missing token body")
	}

	x.version = body.GetVersion()

	binID := body.GetId()
	if x.idSet = len(binID) > 0; x.idSet {
		err := x.id.UnmarshalBinary(binID)
		if err != nil {
			return fmt.Errorf("invalid session ID: %w", err)
		}
		if ver := x.id.Version(); ver != 4 {
			return fmt.Errorf("invalid session ID: wrong UUID version %d, expected 4", ver)
		}
	} else if checkFieldPresence {
		return errors.New("missing session ID")
	}

	issuer := body.GetIssuer()
	if issuer != nil {
		if err := x.issuer.fromProtoMessage(issuer); err != nil {
			return fmt.Errorf("invalid issuer: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing issuer")
	}

	subjects := body.GetSubjects()
	if len(subjects) > 0 {
		x.subjects = make([]Target, len(subjects))
		for i, subj := range subjects {
			if err := x.subjects[i].fromProtoMessage(subj); err != nil {
				return fmt.Errorf("invalid subject at index %d: %w", i, err)
			}
		}
	} else if checkFieldPresence {
		return errors.New("missing subjects")
	}

	lifetime := body.GetLifetime()
	if checkFieldPresence && lifetime == nil {
		return errors.New("missing token lifetime")
	}
	if lifetime != nil {
		x.Lifetime = NewLifetime(
			lifetime.GetIat(),
			lifetime.GetNbf(),
			lifetime.GetExp(),
		)
	}

	contexts := body.GetContexts()
	if len(contexts) > 0 {
		x.contexts = make([]ContextV2, len(contexts))
		for i, ctx := range contexts {
			if err := x.contexts[i].fromProtoMessage(ctx); err != nil {
				return fmt.Errorf("invalid context at index %d: %w", i, err)
			}
		}
	} else if checkFieldPresence {
		return errors.New("missing contexts")
	}

	chain := m.GetDelegationChain()
	if len(chain) > 0 {
		x.delegationChain = make([]DelegationInfo, len(chain))
		for i, del := range chain {
			if err := x.delegationChain[i].fromProtoMessage(del); err != nil {
				return fmt.Errorf("invalid delegation at index %d: %w", i, err)
			}
		}
	}

	if x.sigSet = m.Signature != nil; x.sigSet {
		if err := x.sig.FromProtoMessage(m.Signature); err != nil {
			return fmt.Errorf("invalid body signature: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing body signature")
	}

	return nil
}

func (x TokenV2) fillBody() *protosession.SessionTokenV2_Body {
	body := &protosession.SessionTokenV2_Body{
		Version: x.version,
	}

	if x.idSet {
		body.Id = x.id[:]
	}

	if x.issuer.IsOwnerID() || x.issuer.IsNNS() {
		body.Issuer = x.issuer.protoMessage()
	}

	if len(x.subjects) > 0 {
		body.Subjects = make([]*protosession.Target, len(x.subjects))
		for i := range x.subjects {
			body.Subjects[i] = x.subjects[i].protoMessage()
		}
	}

	if x.Iat() != 0 || x.Nbf() != 0 || x.Exp() != 0 {
		body.Lifetime = &protosession.TokenLifetime{
			Iat: x.Iat(),
			Nbf: x.Nbf(),
			Exp: x.Exp(),
		}
	}

	if len(x.contexts) > 0 {
		body.Contexts = make([]*protosession.SessionContextV2, len(x.contexts))
		for i := range x.contexts {
			body.Contexts[i] = x.contexts[i].protoMessage()
		}
	}

	return body
}

// ProtoMessage converts x into message to transmit using the NeoFS API protocol.
func (x TokenV2) ProtoMessage() *protosession.SessionTokenV2 {
	m := &protosession.SessionTokenV2{
		Body: x.fillBody(),
	}

	if x.sigSet {
		m.Signature = x.sig.ProtoMessage()
	}

	if len(x.delegationChain) > 0 {
		m.DelegationChain = make([]*protosession.DelegationInfo, len(x.delegationChain))
		for i := range x.delegationChain {
			m.DelegationChain[i] = x.delegationChain[i].protoMessage()
		}
	}

	return m
}

// SignedData returns actual payload to sign.
func (x TokenV2) SignedData() []byte {
	return neofsproto.MarshalMessage(x.fillBody())
}

// Sign calculates and writes signature of the TokenV2 data using signer.
func (x *TokenV2) Sign(signer user.Signer) error {
	x.issuer = NewTarget(signer.UserID())
	if x.issuer.IsEmpty() {
		return user.ErrZeroID
	}
	x.sigSet = true
	return x.sig.Calculate(signer, x.SignedData())
}

// AttachSignature attaches given signature to the token.
func (x *TokenV2) AttachSignature(sig neofscrypto.Signature) {
	x.sig, x.sigSet = sig, true
}

// Signature returns token signature.
func (x TokenV2) Signature() (neofscrypto.Signature, bool) {
	return x.sig, x.sigSet
}

// VerifySignature checks if TokenV2 signature is presented and valid.
func (x TokenV2) VerifySignature() bool {
	return x.sigSet && x.sig.Verify(x.SignedData())
}

// Marshal encodes TokenV2 into a binary format of the NeoFS API protocol
// (Protocol Buffers with direct field order).
func (x TokenV2) Marshal() []byte {
	return neofsproto.MarshalMessage(x.ProtoMessage())
}

// Unmarshal decodes NeoFS API protocol binary format into the TokenV2
// (Protocol Buffers with direct field order).
func (x *TokenV2) Unmarshal(data []byte) error {
	var m protosession.SessionTokenV2
	err := neofsproto.UnmarshalMessage(data, &m)
	if err != nil {
		return err
	}
	return x.fromProtoMessage(&m, false)
}

// MarshalJSON encodes TokenV2 into a JSON format of the NeoFS API protocol
// (Protocol Buffers JSON).
func (x TokenV2) MarshalJSON() ([]byte, error) {
	return neofsproto.MarshalMessageJSON(x.ProtoMessage())
}

// UnmarshalJSON decodes NeoFS API protocol JSON format into the TokenV2
// (Protocol Buffers JSON).
func (x *TokenV2) UnmarshalJSON(data []byte) error {
	var m protosession.SessionTokenV2
	err := neofsproto.UnmarshalMessageJSON(data, &m)
	if err != nil {
		return err
	}
	return x.fromProtoMessage(&m, false)
}

// Validate performs comprehensive validation of the token.
func (x TokenV2) Validate() error {
	if x.version != TokenV2CurrentVersion {
		return fmt.Errorf("invalid token version: expected %d, got %d", TokenV2CurrentVersion, x.version)
	}

	if !x.idSet || x.id == uuid.Nil {
		return errors.New("token ID is not set")
	}

	if x.issuer.IsEmpty() {
		return errors.New("issuer is not set")
	}

	if len(x.subjects) == 0 {
		return errors.New("no subjects specified")
	}

	for i, subj := range x.subjects {
		if subj.IsEmpty() {
			return fmt.Errorf("subject at index %d is empty", i)
		}
	}

	if x.Iat() == 0 {
		return errors.New("issued at (iat) is not set")
	}
	if x.Nbf() == 0 {
		return errors.New("not before (nbf) is not set")
	}
	if x.Exp() == 0 {
		return errors.New("expiration (exp) is not set")
	}
	if x.Nbf() > x.Exp() {
		return errors.New("not before (nbf) is after expiration (exp)")
	}
	if x.Iat() > x.Exp() {
		return errors.New("issued at (iat) is after expiration (exp)")
	}

	if len(x.contexts) == 0 {
		return errors.New("no contexts specified")
	}
	for i, ctx := range x.contexts {
		if len(ctx.verbs) == 0 {
			return fmt.Errorf("context at index %d has no verbs", i)
		}
	}

	if err := x.ValidateDelegationChain(); err != nil {
		return fmt.Errorf("invalid delegation chain: %w", err)
	}

	if !x.sigSet {
		return errors.New("token is not signed")
	}
	if !x.VerifySignature() {
		return errors.New("token signature verification failed")
	}

	return nil
}

// ValidateDelegationChain verifies the entire delegation chain is valid.
func (x TokenV2) ValidateDelegationChain() error {
	if len(x.delegationChain) == 0 {
		return nil // Empty chain is valid
	}

	// First delegation must start from token issuer
	if !x.delegationChain[0].issuer.Equals(x.issuer) {
		return errors.New("delegation chain doesn't start from token issuer")
	}

	// Track available verbs as we traverse the chain
	availableVerbs := make(map[VerbV2]bool)
	for _, ctx := range x.contexts {
		for _, v := range ctx.verbs {
			availableVerbs[v] = true
		}
	}

	for i := range x.delegationChain {
		del := &x.delegationChain[i]

		if !del.sigSet {
			return fmt.Errorf("delegation %d is not signed", i)
		}
		if !del.VerifySignature() {
			return fmt.Errorf("delegation %d has invalid signature", i)
		}

		if del.issuer.IsEmpty() {
			return fmt.Errorf("delegation %d has empty issuer", i)
		}
		if len(del.subjects) == 0 {
			return fmt.Errorf("delegation %d has no subjects", i)
		}
		for j, subj := range del.subjects {
			if subj.IsEmpty() {
				return fmt.Errorf("delegation %d has empty subject at index %d", i, j)
			}
		}

		// Check chain continuity - next delegation's issuer must be one of current delegation's subjects
		if i > 0 {
			prevSubjects := x.delegationChain[i-1].subjects
			found := false
			for _, prevSubj := range prevSubjects {
				if x.delegationChain[i].issuer.Equals(prevSubj) {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("delegation chain broken at index %d: issuer doesn't match any previous subject", i)
			}
		}

		// Check delegation lifetime is within token lifetime and within previous delegation lifetime
		if del.Nbf() < x.Nbf() || del.Exp() > x.Exp() {
			return fmt.Errorf("delegation %d lifetime is outside token lifetime", i)
		}

		if i > 0 {
			prevDel := &x.delegationChain[i-1]
			if del.Nbf() < prevDel.Nbf() || del.Exp() > prevDel.Exp() {
				return fmt.Errorf("delegation %d lifetime extends beyond previous delegation lifetime", i)
			}
		}

		// Verify delegated verbs don't exceed available verbs (principle of least privilege)
		newAvailableVerbs := make(map[VerbV2]bool)
		for _, verb := range del.verbs {
			if !availableVerbs[verb] {
				return fmt.Errorf("delegation %d tries to delegate verb %d which was not available", i, verb)
			}
			newAvailableVerbs[verb] = true
		}
		availableVerbs = newAvailableVerbs
	}

	// The first delegation's subjects must be in token subjects
	// This is who the issuer directly authorized
	for _, firstSubj := range x.delegationChain[0].subjects {
		found := false
		for _, subj := range x.subjects {
			if firstSubj.Equals(subj) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("first delegation subject %v is not in token subjects", firstSubj)
		}
	}

	return nil
}

// AssertAuthority checks if the given target is authorized by this token.
// This includes checking if the target is in the subjects list or in the delegation chain.
// When checking delegation chain, it validates the entire chain is valid and properly connects to the target.
func (x TokenV2) AssertAuthority(target Target) bool {
	// Check if target is directly authorized
	for _, subj := range x.subjects {
		if subj.Equals(target) {
			return true
		}
	}

	if len(x.delegationChain) > 0 {
		if err := x.ValidateDelegationChain(); err != nil {
			return false
		}

		// Check if target is in the delegation chain
		for _, del := range x.delegationChain {
			for _, subj := range del.subjects {
				if subj.Equals(target) {
					return true
				}
			}
		}
	}

	return false
}

// AssertVerb checks if the given verb is authorized in any of the contexts.
func (x TokenV2) AssertVerb(verb VerbV2, container cid.ID) bool {
	for _, ctx := range x.contexts {
		if !ctx.container.IsZero() && ctx.container != container {
			continue
		}

		for _, v := range ctx.verbs {
			if v == verb {
				return true
			}
		}
	}

	return false
}

// AssertObject checks if operations on the given object are authorized.
func (x TokenV2) AssertObject(verb VerbV2, container cid.ID, object oid.ID) bool {
	if !verb.IsObjectVerb() {
		return false
	}

	for _, ctx := range x.contexts {
		if !ctx.container.IsZero() && ctx.container != container {
			continue
		}

		verbAuthorized := false
		for _, v := range ctx.verbs {
			if v == verb {
				verbAuthorized = true
				break
			}
		}
		if !verbAuthorized {
			continue
		}

		if len(ctx.objects) == 0 {
			return true
		}

		for _, obj := range ctx.objects {
			if obj == object {
				return true
			}
		}
	}

	return false
}

// AssertContainer checks if container operations with container verbs are authorized for the given container.
// This method specifically checks for container-level operations (ContainerPut, ContainerDelete, ContainerSetEACL).
func (x TokenV2) AssertContainer(verb VerbV2, container cid.ID) bool {
	if !verb.IsContainerVerb() {
		return false
	}

	for _, ctx := range x.contexts {
		if !ctx.container.IsZero() && ctx.container != container {
			continue
		}

		for _, v := range ctx.verbs {
			if v == verb {
				return true
			}
		}
	}

	return false
}
