package session

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"slices"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// TokenCurrentVersion is the current [Token] version.
const TokenCurrentVersion = 0

// RandomNonce generates a random nonce value for use in session tokens.
// Panics if the random number generation fails (which should never happen in practice).
func RandomNonce() uint32 {
	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic(fmt.Errorf("failed to generate random nonce: %w", err))
	}
	return binary.BigEndian.Uint32(buf[:])
}

// Verb represents all possible operations in NeoFS
// that can be authorized via session tokens or delegation chains.
type Verb int32

const (
	VerbUnspecified Verb = iota
	VerbObjectPut
	VerbObjectGet
	VerbObjectHead
	VerbObjectSearch
	VerbObjectDelete
	VerbObjectRange
	VerbObjectRangeHash
	VerbContainerPut
	VerbContainerDelete
	VerbContainerSetEACL
)

// Token repeated field limits.
const (
	// MaxSubjectsPerToken is the maximum number of subjects allowed in a Token.
	MaxSubjectsPerToken = 8
	// MaxContextsPerToken is the maximum number of contexts allowed in a Token.
	MaxContextsPerToken = 16
	// MaxVerbsPerContext is the maximum number of verbs allowed in a SessionContextV2.
	MaxVerbsPerContext = 10
	// MaxObjectsPerContext is the maximum number of objects allowed in a SessionContextV2.
	MaxObjectsPerContext = 1000
	// MaxDelegationDepth is the maximum depth of the delegation chain.
	MaxDelegationDepth = 4
)

// Lifetime represents token or delegation lifetime claims.
type Lifetime struct {
	iat uint64
	nbf uint64
	exp uint64
}

// NewLifetime creates a new Lifetime instance.
// Parameters iat, nbf, exp should be Unix timestamps in seconds.
func NewLifetime(iat, nbf, exp uint64) Lifetime {
	return Lifetime{
		iat: iat,
		nbf: nbf,
		exp: exp,
	}
}

// Iat returns the `issued at` claim as a Unix timestamp.
func (l Lifetime) Iat() uint64 {
	return l.iat
}

// Nbf returns the `not valid before` claim as a Unix timestamp.
func (l Lifetime) Nbf() uint64 {
	return l.nbf
}

// Exp returns the `expiration` claim as a Unix timestamp.
func (l Lifetime) Exp() uint64 {
	return l.exp
}

// SetIat sets the `issued at` claim as a Unix timestamp.
func (l *Lifetime) SetIat(iat uint64) {
	l.iat = iat
}

// SetNbf sets the `not valid before` claim as a Unix timestamp.
func (l *Lifetime) SetNbf(nbf uint64) {
	l.nbf = nbf
}

// SetExp sets the `expiration` claim as a Unix timestamp.
func (l *Lifetime) SetExp(exp uint64) {
	l.exp = exp
}

// ValidAt checks if the lifetime is valid at the given Unix timestamp.
func (l Lifetime) ValidAt(timestamp uint64) bool {
	return timestamp >= l.iat && timestamp <= l.exp && timestamp >= l.nbf
}

// Target represents an account identifier in session token V2.
// Target implements built-in comparable interface.
type Target struct {
	ownerID user.ID
	nnsName string
}

// NewTargetUser creates a new Target from user.ID.
func NewTargetUser(id user.ID) Target {
	return Target{ownerID: id}
}

// NewTargetNamed creates a new Target from NNS name.
func NewTargetNamed(nnsName string) Target {
	return Target{nnsName: nnsName}
}

// IsUserID returns true if target uses direct UserID reference.
func (t Target) IsUserID() bool {
	return !t.ownerID.IsZero()
}

// IsNNS returns true if target uses NNS name reference.
func (t Target) IsNNS() bool {
	return t.nnsName != ""
}

// UserID returns the user ID if target uses direct reference.
func (t Target) UserID() user.ID {
	return t.ownerID
}

// NNSName returns the NNS name if target uses NNS reference.
func (t Target) NNSName() string {
	return t.nnsName
}

// String returns string representation of the target.
func (t Target) String() string {
	if t.IsUserID() {
		return t.ownerID.String()
	}
	return t.nnsName
}

// IsEmpty checks if the target is empty.
func (t Target) IsEmpty() bool {
	return t.ownerID.IsZero() && t.nnsName == ""
}

func (t Target) protoMessage() *protosession.Target {
	switch {
	case t.IsUserID():
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

// Context represents a unified session context for both object and container operations.
// Limits session permissions to specific containers, objects, and operations.
type Context struct {
	container cid.ID
	objects   []oid.ID
	verbs     []Verb
}

// NewContext creates a new unified session context.
// The input verbs slice must not be modified after the call.
// Returns error if the number of verbs exceeds MaxVerbsPerContext or is zero.
func NewContext(container cid.ID, verbs []Verb) (Context, error) {
	if len(verbs) == 0 {
		return Context{}, errors.New("no verbs specified")
	}
	if len(verbs) > MaxVerbsPerContext {
		return Context{}, fmt.Errorf("too many verbs: expected max %d, got %d", MaxVerbsPerContext, len(verbs))
	}
	return Context{
		container: container,
		verbs:     verbs,
	}, nil
}

// Container returns the container ID for this context.
func (c Context) Container() cid.ID {
	return c.container
}

// Objects returns the specific objects for this context.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (c Context) Objects() []oid.ID {
	return c.objects
}

// Verbs returns the authorized operations for this context.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (c Context) Verbs() []Verb {
	return c.verbs
}

// SetObjects sets specific objects for this context.
// The input objects slice must not be modified after the call.
// Returns error if the number of objects exceeds MaxObjectsPerContext.
func (c *Context) SetObjects(objects []oid.ID) error {
	if len(objects) > MaxObjectsPerContext {
		return fmt.Errorf("too many objects: expected max %d, got %d", MaxObjectsPerContext, len(objects))
	}
	c.objects = objects
	return nil
}

func (c Context) protoMessage() *protosession.SessionContextV2 {
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

func (c *Context) fromProtoMessage(m *protosession.SessionContextV2) error {
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

	c.verbs = make([]Verb, len(m.Verbs))
	for i, v := range m.Verbs {
		if v < 0 {
			return fmt.Errorf("negative verb %d", v)
		}
		c.verbs[i] = Verb(v)
	}

	return nil
}

// Token represents NeoFS Session Token V2 with enhanced capabilities:
// - Multiple subjects (authorized users)
// - Multiple contexts (combined container and object operations)
// - Multiple verbs per context
// - Delegation chain tracking
// - NNS name resolution support
//
// Token is mutually compatible with [protosession.SessionTokenV2] message.
// See [Token.FromProtoMessage] / [Token.ProtoMessage] methods.
//
// Instances can be created using built-in var declaration.
type Token struct {
	Lifetime
	version  uint32
	nonce    uint32
	issuer   user.ID
	subjects []Target
	contexts []Context
	final    bool

	origin *Token

	sigSet bool
	sig    neofscrypto.Signature
}

// CopyTo writes deep copy of the [Token] to dst.
func (x Token) CopyTo(dst *Token) {
	dst.version = x.version
	dst.nonce = x.nonce
	dst.issuer = x.issuer
	dst.subjects = slices.Clone(x.subjects)
	dst.Lifetime = x.Lifetime
	dst.contexts = slices.Clone(x.contexts)
	dst.final = x.final
	if x.origin != nil {
		originCopy := new(Token)
		x.origin.CopyTo(originCopy)
		dst.origin = originCopy
	} else {
		dst.origin = nil
	}
	dst.sigSet = x.sigSet
	if x.sigSet {
		dst.sig = neofscrypto.NewSignatureFromRawKey(x.sig.Scheme(), bytes.Clone(x.sig.PublicKeyBytes()), bytes.Clone(x.sig.Value()))
	} else {
		dst.sig = neofscrypto.Signature{}
	}
}

// SetVersion sets the token version.
func (x *Token) SetVersion(version uint32) {
	x.version = version
}

// Version returns the token version.
func (x Token) Version() uint32 {
	return x.version
}

// SetNonce sets the token nonce to prevent collision
// of tokens with the same fields.
func (x *Token) SetNonce(nonce uint32) {
	x.nonce = nonce
}

// Nonce returns the token nonce.
func (x Token) Nonce() uint32 {
	return x.nonce
}

// SetIssuer allows to set issuer before Sign call.
// Using this method is not required when Sign is used (issuer will be derived from the signer automatically).
// When using it please ensure that the token is signed with the same signer as the issuer passed here.
func (x *Token) SetIssuer(issuer user.ID) {
	x.issuer = issuer
}

// Issuer returns the account that issued this token.
func (x Token) Issuer() user.ID {
	return x.issuer
}

// OriginalIssuer returns the account that issued the original token in the delegation chain.
// If the token has no origin, it returns the issuer of this token.
func (x Token) OriginalIssuer() user.ID {
	if x.origin == nil {
		return x.issuer
	}
	return x.origin.OriginalIssuer()
}

// SetSubjects sets the accounts authorized by this token.
// The input subjects slice must not be modified after the call.
// Returns error if the number of subjects exceeds MaxSubjectsPerToken.
func (x *Token) SetSubjects(subjects []Target) error {
	if len(subjects) > MaxSubjectsPerToken {
		return fmt.Errorf("too many subjects: expected max %d, got %d", MaxSubjectsPerToken, len(subjects))
	}
	x.subjects = subjects
	return nil
}

// Subjects returns the accounts authorized by this token.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (x Token) Subjects() []Target {
	return x.subjects
}

// AddSubject adds an authorized account to this token.
// Returns error if adding the subject would exceed MaxSubjectsPerToken.
func (x *Token) AddSubject(subject Target) error {
	if len(x.subjects) >= MaxSubjectsPerToken {
		return fmt.Errorf("cannot add subject: already at maximum of %d", MaxSubjectsPerToken)
	}
	x.subjects = append(x.subjects, subject)
	return nil
}

// SetContexts sets the session contexts.
// The input contexts slice must not be modified after the call.
// Returns error if the number of contexts exceeds MaxContextsPerToken.
func (x *Token) SetContexts(contexts []Context) error {
	if len(contexts) > MaxContextsPerToken {
		return fmt.Errorf("too many contexts: expected max %d, got %d", MaxContextsPerToken, len(contexts))
	}
	x.contexts = contexts
	return nil
}

// Contexts returns the session contexts.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (x Token) Contexts() []Context {
	return x.contexts
}

// AddContext adds a session context.
// Returns error if adding the context would exceed MaxContextsPerToken.
func (x *Token) AddContext(ctx Context) error {
	if len(x.contexts) >= MaxContextsPerToken {
		return fmt.Errorf("cannot add context: already at maximum of %d", MaxContextsPerToken)
	}
	x.contexts = append(x.contexts, ctx)
	return nil
}

// SetFinal marks the token as final, preventing further delegations.
func (x *Token) SetFinal(final bool) {
	x.final = final
}

// IsFinal returns true if the token is marked as final.
// Final tokens cannot be further delegated.
func (x Token) IsFinal() bool {
	return x.final
}

// SetOrigin sets the origin token for delegation chain.
// This creates a link to the token that was delegated to create this token.
func (x *Token) SetOrigin(origin *Token) {
	x.origin = origin
}

// Origin returns the origin token if this token is part of a delegation chain.
func (x Token) Origin() *Token {
	return x.origin
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// x from it.
func (x *Token) FromProtoMessage(m *protosession.SessionTokenV2) error {
	return x.fromProtoMessage(m, true)
}

func (x *Token) fromProtoMessage(m *protosession.SessionTokenV2, checkFieldPresence bool) error {
	body := m.GetBody()
	if checkFieldPresence && body == nil {
		return errors.New("missing token body")
	}

	x.version = body.GetVersion()
	x.nonce = body.GetNonce()
	x.final = body.GetFinal()

	issuer := body.GetIssuer()
	if issuer != nil {
		if err := x.issuer.FromProtoMessage(issuer); err != nil {
			return fmt.Errorf("invalid issuer: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing issuer")
	} else {
		x.issuer = user.ID{}
	}

	subjects := body.GetSubjects()
	if len(subjects) > 0 {
		x.subjects = make([]Target, len(subjects))
		for i, subj := range subjects {
			if subj == nil {
				return fmt.Errorf("nil subject at index %d", i)
			}
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
		x.contexts = make([]Context, len(contexts))
		for i, ctx := range contexts {
			if ctx == nil {
				return fmt.Errorf("nil context at index %d", i)
			}
			if err := x.contexts[i].fromProtoMessage(ctx); err != nil {
				return fmt.Errorf("invalid context at index %d: %w", i, err)
			}
		}
	} else if checkFieldPresence {
		return errors.New("missing contexts")
	}

	if m.Origin != nil {
		x.origin = &Token{}
		if err := x.origin.fromProtoMessage(m.Origin, checkFieldPresence); err != nil {
			return fmt.Errorf("invalid origin token: %w", err)
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

func (x Token) fillBody() *protosession.SessionTokenV2_Body {
	body := &protosession.SessionTokenV2_Body{
		Version: x.version,
		Nonce:   x.nonce,
		Final:   x.final,
	}

	if !x.issuer.IsZero() {
		body.Issuer = x.issuer.ProtoMessage()
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
func (x Token) ProtoMessage() *protosession.SessionTokenV2 {
	m := &protosession.SessionTokenV2{
		Body: x.fillBody(),
	}

	if x.sigSet {
		m.Signature = x.sig.ProtoMessage()
	}

	if x.origin != nil {
		m.Origin = x.origin.ProtoMessage()
	}

	return m
}

// SignedData returns actual payload to sign.
func (x Token) SignedData() []byte {
	return neofsproto.MarshalMessage(x.fillBody())
}

// Sign calculates and writes signature of the Token data using signer.
func (x *Token) Sign(signer user.Signer) error {
	x.issuer = signer.UserID()
	if x.issuer.IsZero() {
		return user.ErrZeroID
	}
	x.sigSet = true
	return x.sig.Calculate(signer, x.SignedData())
}

// AttachSignature attaches given signature to the token.
func (x *Token) AttachSignature(sig neofscrypto.Signature) {
	x.sig, x.sigSet = sig, true
}

// Signature returns token signature.
func (x Token) Signature() (neofscrypto.Signature, bool) {
	return x.sig, x.sigSet
}

// VerifySignature checks if Token signature is presented and valid.
func (x Token) VerifySignature() bool {
	return x.sigSet && x.sig.Verify(x.SignedData())
}

// Marshal encodes Token into a binary format of the NeoFS API protocol
// (Protocol Buffers with direct field order).
func (x Token) Marshal() []byte {
	return neofsproto.MarshalMessage(x.ProtoMessage())
}

// Unmarshal decodes NeoFS API protocol binary format into the Token
// (Protocol Buffers with direct field order).
func (x *Token) Unmarshal(data []byte) error {
	var m protosession.SessionTokenV2
	err := neofsproto.UnmarshalMessage(data, &m)
	if err != nil {
		return err
	}
	return x.fromProtoMessage(&m, false)
}

// MarshalJSON encodes Token into a JSON format of the NeoFS API protocol
// (Protocol Buffers JSON).
func (x Token) MarshalJSON() ([]byte, error) {
	return neofsproto.MarshalMessageJSON(x.ProtoMessage())
}

// UnmarshalJSON decodes NeoFS API protocol JSON format into the Token
// (Protocol Buffers JSON).
func (x *Token) UnmarshalJSON(data []byte) error {
	var m protosession.SessionTokenV2
	err := neofsproto.UnmarshalMessageJSON(data, &m)
	if err != nil {
		return err
	}
	return x.fromProtoMessage(&m, false)
}

// Validate performs comprehensive validation of the token.
// It includes field checks, depth checks, delegation chain validation,
// narrowing of authorized verbs and lifetimes and issuer authorization checks.
// The chain is built using recursive origin tokens, with a maximum depth of 4.
// NNS subjects are not considered during issuer authorization checks.
// If NNS resolution is needed, use ValidateWithNNS instead.
func (x Token) Validate() error {
	return x.ValidateWithNNS(nil)
}

// ValidateWithNNS performs comprehensive validation of the token.
// It includes field checks, depth checks, delegation chain validation,
// narrowing of authorized verbs and lifetimes and issuer authorization checks.
// The chain is built using recursive origin tokens, with a maximum depth of 4.
// The hasUserInNNS function checks if a user ID is authorized under a given NNS name.
// If no NNS resolution is nil, NNS subjects are not considered during issuer authorization checks.
func (x Token) ValidateWithNNS(hasUserInNNS func(name string, userID user.ID) bool) error {
	seen := make(map[*Token]bool)
	if err := x.validate(0, seen, hasUserInNNS); err != nil {
		return err
	}
	return nil
}

// validate recursively validates the token and its delegation chain.
func (x *Token) validate(depth int, seen map[*Token]bool, hasUserInNNS func(string, user.ID) bool) error {
	// Depth check
	if seen[x] {
		return errors.New("cycle detected in delegation chain")
	}
	seen[x] = true
	if depth > MaxDelegationDepth {
		return fmt.Errorf("delegation chain exceeds maximum depth of %d", MaxDelegationDepth)
	}

	prefix := ""
	if depth > 0 {
		prefix = fmt.Sprintf("depth %d: ", depth)
	}
	// Field validation
	if err := x.validateFields(); err != nil {
		return fmt.Errorf("%sinvalid fields: %w", prefix, err)
	}

	// Final tokens cannot be embedded in other tokens (cannot be further delegated)
	if x.IsFinal() && depth > 0 {
		return errors.New(prefix + "final token cannot be used as origin (further delegated)")
	}

	// No origin means no further validation needed
	if x.origin == nil {
		return nil
	}

	// Delegated verbs validation
	if err := x.validateDelegatedVerbs(); err != nil {
		return fmt.Errorf("%sinvalid origin chain: %w", prefix, err)
	}

	// Issuer authorization check
	foundIssuer := false
	for _, subj := range x.origin.subjects {
		if subj.IsUserID() && x.issuer == subj.UserID() {
			foundIssuer = true
			break
		}
		if subj.IsNNS() && hasUserInNNS != nil {
			if hasUserInNNS(subj.NNSName(), x.issuer) {
				foundIssuer = true
				break
			}
		}
	}
	if !foundIssuer && hasUserInNNS == nil {
		return errors.New(prefix + "token issuer is not in this origin token's subjects")
	}

	// Lifetime narrowing check
	if x.origin.Nbf() > x.Nbf() || x.origin.Exp() < x.Exp() {
		return errors.New(prefix + "origin token lifetime is outside this token's lifetime")
	}

	depth++
	return x.origin.validate(depth, seen, hasUserInNNS)
}

// validateFields checks individual fields of the token for validity.
func (x Token) validateFields() error {
	if x.version != TokenCurrentVersion {
		return fmt.Errorf("invalid token version: expected %d, got %d", TokenCurrentVersion, x.version)
	}

	if x.issuer.IsZero() {
		return errors.New("issuer is not set")
	}

	if len(x.subjects) == 0 {
		return errors.New("no subjects specified")
	}

	if len(x.subjects) > MaxSubjectsPerToken {
		return fmt.Errorf("too many subjects: expected max %d, got %d", MaxSubjectsPerToken, len(x.subjects))
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

	if len(x.contexts) > MaxContextsPerToken {
		return fmt.Errorf("too many contexts: expected max %d, got %d", MaxContextsPerToken, len(x.contexts))
	}

	for i, ctx := range x.contexts {
		if len(ctx.verbs) == 0 {
			return fmt.Errorf("context at index %d has no verbs", i)
		}
		if len(ctx.verbs) > MaxVerbsPerContext {
			return fmt.Errorf("context at index %d has too many verbs: expected max %d, got %d", i, MaxVerbsPerContext, len(ctx.verbs))
		}
		if len(ctx.objects) > MaxObjectsPerContext {
			return fmt.Errorf("context at index %d has too many objects: expected max %d, got %d", i, MaxObjectsPerContext, len(ctx.objects))
		}
	}

	if !x.sigSet {
		return errors.New("token is not signed")
	}
	if !x.VerifySignature() {
		return errors.New("token signature verification failed")
	}

	return nil
}

// validateDelegatedVerbs checks that all verbs in this token are authorized by the origin token.
// Each context in the delegated token must have corresponding authorization in the origin.
func (x Token) validateDelegatedVerbs() error {
	for i, ctx := range x.contexts {
		authorized := false

		for _, originCtx := range x.origin.contexts {
			containerMatches := originCtx.container.IsZero() || originCtx.container == ctx.container
			if !containerMatches {
				continue
			}

			allVerbsAuthorized := true
			for _, verb := range ctx.verbs {
				verbAuthorized := false
				for _, originVerb := range originCtx.verbs {
					if verb == originVerb {
						verbAuthorized = true
						break
					}
				}
				if !verbAuthorized {
					allVerbsAuthorized = false
					break
				}
			}

			if allVerbsAuthorized {
				authorized = true
				break
			}
		}

		if !authorized {
			return fmt.Errorf("context at index %d uses verbs not authorized by origin token", i)
		}
	}

	return nil
}

// AssertAuthority checks if the given target is authorized by this token's subjects.
// For NNS subjects, it uses the provided hasUserInNNS function to resolve the mapping.
// If hasUserInNNS is nil, NNS subjects are ignored and only direct user ID matches are considered.
func (x Token) AssertAuthority(userID user.ID, hasUserInNNS func(name string, userID user.ID) bool) bool {
	for _, subj := range x.subjects {
		if subj.IsUserID() && subj.UserID() == userID {
			return true
		}

		if subj.IsNNS() && hasUserInNNS != nil {
			if hasUserInNNS(subj.NNSName(), userID) {
				return true
			}
		}
	}

	return false
}

// AssertVerb checks if the given verb is authorized in any of the contexts.
func (x Token) AssertVerb(verb Verb, container cid.ID) bool {
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
func (x Token) AssertObject(verb Verb, container cid.ID, object oid.ID) bool {
	switch verb {
	case VerbObjectPut, VerbObjectGet, VerbObjectHead, VerbObjectSearch,
		VerbObjectDelete, VerbObjectRange, VerbObjectRangeHash:
	default:
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
// This method specifically checks for container-level operations
// ([VerbContainerPut], [VerbContainerDelete], [VerbContainerSetEACL]).
func (x Token) AssertContainer(verb Verb, container cid.ID) bool {
	switch verb {
	case VerbContainerPut, VerbContainerDelete, VerbContainerSetEACL:
	default:
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
