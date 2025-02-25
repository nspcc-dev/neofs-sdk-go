package netmap

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"

	"github.com/antlr4-go/antlr/v4"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	"github.com/nspcc-dev/neofs-sdk-go/netmap/parser"
	protonetmap "github.com/nspcc-dev/neofs-sdk-go/proto/netmap"
)

const defaultContainerBackupFactor = 3

const (
	maxObjectReplicasPerSet = 8
	maxContainerNodesInSet  = 64
	maxContainerNodeSets    = 256
	maxContainerNodes       = 512
)

// PlacementPolicy declares policy to store objects in the NeoFS container.
// Within itself, PlacementPolicy represents a set of rules to select a subset
// of nodes from NeoFS network map - node-candidates for object storage.
//
// PlacementPolicy is mutually compatible with [protonetmap.PlacementPolicy]
// message. See [PlacementPolicy.FromProtoMessage] / [PlacementPolicy.ProtoMessage] methods.
//
// Instances can be created using built-in var declaration.
type PlacementPolicy struct {
	backupFactor uint32

	filters []Filter

	selectors []Selector

	replicas []ReplicaDescriptor
}

// FilterOp defines the matching property.
type FilterOp int32

// Supported FilterOp values.
const (
	_           FilterOp = iota
	FilterOpEQ           // String 'equal'
	FilterOpNE           // String 'not equal'
	FilterOpGT           // Numeric 'greater than'
	FilterOpGE           // Numeric 'greater or equal than'
	FilterOpLT           // Numeric 'less than'
	FilterOpLE           // Numeric 'less or equal than'
	FilterOpOR           // Logical disjunction
	FilterOpAND          // Logical conjunction
)

// String implements [fmt.Stringer].
func (x FilterOp) String() string {
	switch x {
	default:
		return "UNKNOWN"
	case FilterOpEQ:
		return "EQ"
	case FilterOpNE:
		return "NE"
	case FilterOpGT:
		return "GT"
	case FilterOpGE:
		return "GE"
	case FilterOpLT:
		return "LT"
	case FilterOpLE:
		return "LE"
	case FilterOpOR:
		return "OR"
	case FilterOpAND:
		return "AND"
	}
}

func copyFilter(f Filter) Filter {
	filter := f

	if f.subs != nil {
		filter.subs = make([]Filter, len(f.subs))

		for i := range f.subs {
			filter.subs[i] = copyFilter(f.subs[i])
		}
	} else {
		filter.subs = nil
	}

	return filter
}

// CopyTo writes deep copy of the [PlacementPolicy] to dst.
func (p PlacementPolicy) CopyTo(dst *PlacementPolicy) {
	dst.SetContainerBackupFactor(p.backupFactor)

	dst.filters = make([]Filter, len(p.filters))
	for i := range p.filters {
		dst.filters[i] = copyFilter(p.filters[i])
	}

	// protonetmap.Selector is a struct with simple types, no links inside. Just create a new slice and copy all items inside.
	dst.selectors = slices.Clone(p.selectors)

	// protonetmap.Replica is a struct with simple types, no links inside. Just create a new slice and copy all items inside.
	dst.replicas = slices.Clone(p.replicas)
}

func (p *PlacementPolicy) fromProtoMessage(m *protonetmap.PlacementPolicy, checkFieldPresence bool) error {
	if checkFieldPresence && len(m.Replicas) == 0 {
		return errors.New("missing replicas")
	}

	p.replicas = make([]ReplicaDescriptor, len(m.Replicas))
	for i, r := range m.Replicas {
		if r == nil {
			return fmt.Errorf("nil replica #%d", i)
		}
		p.replicas[i].fromProtoMessage(r)
	}

	p.selectors = make([]Selector, len(m.Selectors))
	for i, s := range m.Selectors {
		if s == nil {
			return fmt.Errorf("nil selector #%d", i)
		}
		if err := p.selectors[i].fromProtoMessage(s); err != nil {
			return fmt.Errorf("invalid selector #%d: %w", i, err)
		}
	}

	p.filters = make([]Filter, len(m.Filters))
	for i, f := range m.Filters {
		if f == nil {
			return fmt.Errorf("nil filter #%d", i)
		}
		if err := p.filters[i].fromProtoMessage(f); err != nil {
			return fmt.Errorf("invalid filter #%d: %w", i, err)
		}
	}

	p.backupFactor = m.GetContainerBackupFactor()

	return nil
}

// Marshal encodes PlacementPolicy into a binary format of the NeoFS API
// protocol (Protocol Buffers with direct field order).
//
// See also Unmarshal.
func (p PlacementPolicy) Marshal() []byte {
	return neofsproto.Marshal(p)
}

// Unmarshal decodes NeoFS API protocol binary format into the PlacementPolicy
// (Protocol Buffers with direct field order). Returns an error describing
// a format violation.
//
// See also Marshal.
func (p *PlacementPolicy) Unmarshal(data []byte) error {
	return neofsproto.UnmarshalOptional(data, p, (*PlacementPolicy).fromProtoMessage)
}

// MarshalJSON encodes PlacementPolicy into a JSON format of the NeoFS API
// protocol (Protocol Buffers JSON).
//
// See also UnmarshalJSON.
func (p PlacementPolicy) MarshalJSON() ([]byte, error) {
	return neofsproto.MarshalJSON(p)
}

// UnmarshalJSON decodes NeoFS API protocol JSON format into the PlacementPolicy
// (Protocol Buffers JSON). Returns an error describing a format violation.
//
// See also MarshalJSON.
func (p *PlacementPolicy) UnmarshalJSON(data []byte) error {
	return neofsproto.UnmarshalJSONOptional(data, p, (*PlacementPolicy).fromProtoMessage)
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// p from it.
//
// See also [PlacementPolicy.ProtoMessage].
func (p *PlacementPolicy) FromProtoMessage(m *protonetmap.PlacementPolicy) error {
	return p.fromProtoMessage(m, true)
}

// ProtoMessage converts p into message to transmit using the NeoFS API
// protocol.
//
// See also [PlacementPolicy.FromProtoMessage].
func (p PlacementPolicy) ProtoMessage() *protonetmap.PlacementPolicy {
	m := &protonetmap.PlacementPolicy{
		ContainerBackupFactor: p.backupFactor,
	}
	if len(p.replicas) > 0 {
		m.Replicas = make([]*protonetmap.Replica, len(p.replicas))
		for i := range p.replicas {
			m.Replicas[i] = p.replicas[i].protoMessage()
		}
	}
	if len(p.selectors) > 0 {
		m.Selectors = make([]*protonetmap.Selector, len(p.selectors))
		for i := range p.selectors {
			m.Selectors[i] = p.selectors[i].protoMessage()
		}
	}
	if len(p.filters) > 0 {
		m.Filters = make([]*protonetmap.Filter, len(p.filters))
		for i := range p.filters {
			m.Filters[i] = p.filters[i].protoMessage()
		}
	}
	return m
}

// ReplicaDescriptor replica descriptor characterizes replicas of objects from
// the subset selected by a particular Selector.
type ReplicaDescriptor struct {
	count    uint32
	selector string
}

// fromProtoMessage validates m according to the NeoFS API protocol and restores
// r from it.
func (r *ReplicaDescriptor) fromProtoMessage(m *protonetmap.Replica) {
	r.count = m.Count
	r.selector = m.Selector
}

// protoMessage converts r into message to transmit using the NeoFS API
// protocol.
func (r ReplicaDescriptor) protoMessage() *protonetmap.Replica {
	return &protonetmap.Replica{
		Count:    r.count,
		Selector: r.selector,
	}
}

// SetNumberOfObjects sets number of object replicas.
func (r *ReplicaDescriptor) SetNumberOfObjects(c uint32) {
	r.count = c
}

// NumberOfObjects returns number set using SetNumberOfObjects.
//
// Zero ReplicaDescriptor has zero number of objects.
func (r ReplicaDescriptor) NumberOfObjects() uint32 {
	return r.count
}

// SetSelectorName sets name of the related Selector.
//
// Zero ReplicaDescriptor references to the root bucket's selector: it contains
// all possible nodes to store the object.
//
// See also [ReplicaDescriptor.SelectorName].
func (r *ReplicaDescriptor) SetSelectorName(s string) {
	r.selector = s
}

// SelectorName returns name of the related Selector.
//
// Zero ReplicaDescriptor references to the root bucket's selector: it contains
// all possible nodes to store the object.
//
// See also [ReplicaDescriptor.SetSelectorName].
func (r ReplicaDescriptor) SelectorName() string {
	return r.selector
}

// SetReplicas sets list of object replica's characteristics.
//
// See also [PlacementPolicy.Replicas], [PlacementPolicy.NumberOfReplicas],
// [PlacementPolicy.ReplicaNumberByIndex].
func (p *PlacementPolicy) SetReplicas(rs []ReplicaDescriptor) {
	p.replicas = rs
}

// Replicas returns list of object replica characteristics.
//
// See also [PlacementPolicy.SetReplicas], [PlacementPolicy.NumberOfReplicas],
// [PlacementPolicy.ReplicaNumberByIndex].
func (p PlacementPolicy) Replicas() []ReplicaDescriptor {
	return p.replicas
}

// NumberOfReplicas returns number of replica descriptors set using SetReplicas.
//
// Zero PlacementPolicy has no replicas which is incorrect according to the
// NeoFS API protocol.
func (p PlacementPolicy) NumberOfReplicas() int {
	return len(p.replicas)
}

// ReplicaNumberByIndex returns number of object replicas from the i-th replica
// descriptor. Index MUST be in range [0; NumberOfReplicas()).
//
// Zero PlacementPolicy has no replicas.
func (p PlacementPolicy) ReplicaNumberByIndex(i int) uint32 {
	return p.replicas[i].NumberOfObjects()
}

// SetContainerBackupFactor sets container backup factor: it controls how deep
// NeoFS will search for nodes alternatives to include into container's nodes subset.
//
// Zero PlacementPolicy has zero container backup factor.
//
// See also [PlacementPolicy.ContainerBackupFactor].
func (p *PlacementPolicy) SetContainerBackupFactor(f uint32) {
	p.backupFactor = f
}

// ContainerBackupFactor returns container backup factor: it controls how deep
// NeoFS will search for nodes alternatives to include into container's nodes subset.
//
// Zero PlacementPolicy has zero container backup factor.
//
// See also [PlacementPolicy.SetContainerBackupFactor].
func (p *PlacementPolicy) ContainerBackupFactor() uint32 {
	return p.backupFactor
}

// Selector describes the bucket selection operator: choose a number of nodes
// from the bucket taking the nearest nodes to the related container by hash distance.
type Selector struct {
	name   string
	count  uint32
	clause protonetmap.Clause
	attr   string
	filter string
}

// fromProtoMessage validates m according to the NeoFS API protocol and restores
// s from it.
func (s *Selector) fromProtoMessage(m *protonetmap.Selector) error {
	if m.Clause < 0 {
		return fmt.Errorf("negative clause %d", m.Clause)
	}
	s.name = m.Name
	s.count = m.Count
	s.clause = m.Clause
	s.attr = m.Attribute
	s.filter = m.Filter
	return nil
}

// protoMessage converts s into message to transmit using the NeoFS API
// protocol.
func (s Selector) protoMessage() *protonetmap.Selector {
	return &protonetmap.Selector{
		Name:      s.name,
		Count:     s.count,
		Clause:    s.clause,
		Attribute: s.attr,
		Filter:    s.filter,
	}
}

// SetName sets name with which the Selector can be referenced.
//
// Zero Selector is unnamed.
func (s *Selector) SetName(name string) {
	s.name = name
}

// Name returns name with which the Selector can be referenced.
//
// Zero Selector is unnamed.
//
// See also [Selector.Name].
func (s Selector) Name() string {
	return s.name
}

// SetNumberOfNodes sets number of nodes to select from the bucket.
//
// Zero Selector selects nothing.
//
// See also [Selector.NumberOfNodes].
func (s *Selector) SetNumberOfNodes(num uint32) {
	s.count = num
}

// NumberOfNodes returns number of nodes to select from the bucket.
//
// Zero Selector selects nothing.
//
// See also [Selector.SetNumberOfNodes].
func (s Selector) NumberOfNodes() uint32 {
	return s.count
}

// SelectByBucketAttribute sets attribute of the bucket to select nodes from.
//
// Zero Selector has empty attribute.
//
// See also [Selector.BucketAttribute].
func (s *Selector) SelectByBucketAttribute(bucket string) {
	s.attr = bucket
}

// BucketAttribute returns attribute of the bucket to select nodes from.
//
// Zero Selector has empty attribute.
//
// See also [Selector.SelectByBucketAttribute].
func (s *Selector) BucketAttribute() string {
	return s.attr
}

// SelectSame makes selection algorithm to select only nodes having the same values
// of the bucket attribute.
//
// Zero Selector doesn't specify selection modifier so nodes are selected randomly.
//
// See also [Selector.SelectByBucketAttribute], [Selector.IsSame].
func (s *Selector) SelectSame() {
	s.clause = protonetmap.Clause_SAME
}

// IsSame checks whether selection algorithm is set to select only nodes having
// the same values of the bucket attribute.
//
// See also [Selector.SelectSame].
func (s *Selector) IsSame() bool {
	return s.clause == protonetmap.Clause_SAME
}

// SelectDistinct makes selection algorithm to select only nodes having the different values
// of the bucket attribute.
//
// Zero Selector doesn't specify selection modifier so nodes are selected randomly.
//
// See also [Selector.SelectByBucketAttribute], [Selector.IsDistinct].
func (s *Selector) SelectDistinct() {
	s.clause = protonetmap.Clause_DISTINCT
}

// IsDistinct checks whether selection algorithm is set to select only nodes
// having the different values of the bucket attribute.
//
// See also [Selector.SelectByBucketAttribute], [Selector.SelectDistinct].
func (s *Selector) IsDistinct() bool {
	return s.clause == protonetmap.Clause_DISTINCT
}

// SetFilterName sets reference to pre-filtering nodes for selection.
//
// Zero Selector has no filtering reference.
//
// See also Filter.SetName.
func (s *Selector) SetFilterName(f string) {
	s.filter = f
}

// FilterName returns reference to pre-filtering nodes for selection.
//
// Zero Selector has no filtering reference.
//
// See also [Filter.SetName], [Selector.SetFilterName].
func (s *Selector) FilterName() string {
	return s.filter
}

// SetSelectors sets list of Selector to form the subset of the nodes to store
// container objects.
//
// Zero PlacementPolicy does not declare selectors.
//
// See also [PlacementPolicy.Selectors].
func (p *PlacementPolicy) SetSelectors(ss []Selector) {
	p.selectors = ss
}

// Selectors returns list of Selector to form the subset of the nodes to store
// container objects.
//
// Zero PlacementPolicy does not declare selectors.
//
// See also [PlacementPolicy.SetSelectors].
func (p PlacementPolicy) Selectors() []Selector {
	return p.selectors
}

// Filter contains rules for filtering the node sets.
type Filter struct {
	name string
	key  string
	op   FilterOp
	val  string
	subs []Filter
}

// fromProtoMessage validates m according to the NeoFS API protocol and restores
// x from it.
func (x *Filter) fromProtoMessage(m *protonetmap.Filter) error {
	if m.Op < 0 {
		return fmt.Errorf("negative op %d", m.Op)
	}
	var subs []Filter
	if len(m.Filters) > 0 {
		subs = make([]Filter, len(m.Filters))
		for i := range m.Filters {
			if err := subs[i].fromProtoMessage(m.Filters[i]); err != nil {
				return fmt.Errorf("invalid sub-filter #%d: %w", i, err)
			}
		}
	}
	x.name = m.Name
	x.setAttribute(m.Key, FilterOp(m.Op), m.Value)
	x.subs = subs
	return nil
}

// protoMessage converts x into message to transmit using the NeoFS API
// protocol.
func (x Filter) protoMessage() *protonetmap.Filter {
	m := &protonetmap.Filter{
		Name:  x.name,
		Key:   x.key,
		Op:    protonetmap.Operation(x.op),
		Value: x.val,
	}
	if len(x.subs) > 0 {
		m.Filters = make([]*protonetmap.Filter, len(x.subs))
		for i := range x.subs {
			m.Filters[i] = x.subs[i].protoMessage()
		}
	}
	return m
}

// SetName sets name with which the Filter can be referenced or, for inner filters,
// to which the Filter references. Top-level filters MUST be named. The name
// MUST NOT be '*'.
//
// Zero Filter is unnamed.
//
// See also [Filter.Name].
func (x *Filter) SetName(name string) {
	x.name = name
}

// Name returns name with which the Filter can be referenced or, for inner
// filters, to which the Filter references. Top-level filters MUST be named. The
// name MUST NOT be '*'.
//
// Zero Filter is unnamed.
//
// See also [Filter.SetName].
func (x Filter) Name() string {
	return x.name
}

// Key returns key to the property.
func (x Filter) Key() string {
	return x.key
}

// Op returns operator to match the property.
func (x Filter) Op() FilterOp {
	return x.op
}

// Value returns value to check the property against.
func (x Filter) Value() string {
	return x.val
}

// SubFilters returns list of sub-filters when Filter is complex.
func (x Filter) SubFilters() []Filter {
	return x.subs
}

func (x *Filter) setAttribute(key string, op FilterOp, val string) {
	x.key = key
	x.op = op
	x.val = val
}

// Equal applies the rule to accept only nodes with the same attribute value.
//
// Method SHOULD NOT be called along with other similar methods.
func (x *Filter) Equal(key, value string) {
	x.setAttribute(key, FilterOpEQ, value)
}

// NotEqual applies the rule to accept only nodes with the distinct attribute value.
//
// Method SHOULD NOT be called along with other similar methods.
func (x *Filter) NotEqual(key, value string) {
	x.setAttribute(key, FilterOpNE, value)
}

// NumericGT applies the rule to accept only nodes with the numeric attribute
// greater than given number.
//
// Method SHOULD NOT be called along with other similar methods.
func (x *Filter) NumericGT(key string, num int64) {
	x.setAttribute(key, FilterOpGT, strconv.FormatInt(num, 10))
}

// NumericGE applies the rule to accept only nodes with the numeric attribute
// greater than or equal to given number.
//
// Method SHOULD NOT be called along with other similar methods.
func (x *Filter) NumericGE(key string, num int64) {
	x.setAttribute(key, FilterOpGE, strconv.FormatInt(num, 10))
}

// NumericLT applies the rule to accept only nodes with the numeric attribute
// less than given number.
//
// Method SHOULD NOT be called along with other similar methods.
func (x *Filter) NumericLT(key string, num int64) {
	x.setAttribute(key, FilterOpLT, strconv.FormatInt(num, 10))
}

// NumericLE applies the rule to accept only nodes with the numeric attribute
// less than or equal to given number.
//
// Method SHOULD NOT be called along with other similar methods.
func (x *Filter) NumericLE(key string, num int64) {
	x.setAttribute(key, FilterOpLE, strconv.FormatInt(num, 10))
}

func (x *Filter) setInnerFilters(op FilterOp, filters []Filter) {
	x.setAttribute("", op, "")
	x.subs = filters
}

// LogicalOR applies the rule to accept only nodes which satisfy at least one
// of the given filters.
//
// Method SHOULD NOT be called along with other similar methods.
func (x *Filter) LogicalOR(filters ...Filter) {
	x.setInnerFilters(FilterOpOR, filters)
}

// LogicalAND applies the rule to accept only nodes which satisfy all the given
// filters.
//
// Method SHOULD NOT be called along with other similar methods.
func (x *Filter) LogicalAND(filters ...Filter) {
	x.setInnerFilters(FilterOpAND, filters)
}

// Filters returns list of Filter that will be applied when selecting nodes.
//
// Zero PlacementPolicy has no filters.
//
// See also [PlacementPolicy.SetFilters].
func (p PlacementPolicy) Filters() []Filter {
	return p.filters
}

// SetFilters sets list of Filter that will be applied when selecting nodes.
//
// Zero PlacementPolicy has no filters.
//
// See also [PlacementPolicy.Filters].
func (p *PlacementPolicy) SetFilters(fs []Filter) {
	p.filters = fs
}

// WriteStringTo encodes PlacementPolicy into human-readably query and writes
// the result into w. Returns w's errors directly.
//
// See also DecodeString.
func (p PlacementPolicy) WriteStringTo(w io.StringWriter) (err error) {
	writtenSmth := false

	writeLnIfNeeded := func() error {
		if writtenSmth {
			_, err = w.WriteString("\n")
			return err
		}

		writtenSmth = true

		return nil
	}

	for i := range p.replicas {
		err = writeLnIfNeeded()
		if err != nil {
			return err
		}

		c := p.replicas[i].NumberOfObjects()
		s := p.replicas[i].SelectorName()

		if s != "" {
			_, err = w.WriteString(fmt.Sprintf("REP %d IN %s", c, s))
		} else {
			_, err = w.WriteString(fmt.Sprintf("REP %d", c))
		}

		if err != nil {
			return err
		}
	}

	if p.backupFactor > 0 {
		err = writeLnIfNeeded()
		if err != nil {
			return err
		}

		_, err = w.WriteString(fmt.Sprintf("CBF %d", p.backupFactor))
		if err != nil {
			return err
		}
	}

	var s string

	for i := range p.selectors {
		err = writeLnIfNeeded()
		if err != nil {
			return err
		}

		_, err = w.WriteString(fmt.Sprintf("SELECT %d", p.selectors[i].NumberOfNodes()))
		if err != nil {
			return err
		}

		if s = p.selectors[i].BucketAttribute(); s != "" {
			var clause string

			switch p.selectors[i].clause {
			case protonetmap.Clause_SAME:
				clause = "SAME "
			case protonetmap.Clause_DISTINCT:
				clause = "DISTINCT "
			default:
				clause = ""
			}

			_, err = w.WriteString(fmt.Sprintf(" IN %s%s", clause, s))
			if err != nil {
				return err
			}
		}

		if s = p.selectors[i].FilterName(); s != "" {
			_, err = w.WriteString(" FROM " + s)
			if err != nil {
				return err
			}
		}

		if s = p.selectors[i].Name(); s != "" {
			_, err = w.WriteString(" AS " + s)
			if err != nil {
				return err
			}
		}
	}

	for i := range p.filters {
		err = writeLnIfNeeded()
		if err != nil {
			return err
		}

		_, err = w.WriteString("FILTER ")
		if err != nil {
			return err
		}

		err = writeFilterStringTo(w, p.filters[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func writeFilterStringTo(w io.StringWriter, f Filter) error {
	var err error
	var s string
	op := f.Op()
	unspecified := op == 0

	if s = f.Key(); s != "" {
		_, err = w.WriteString(fmt.Sprintf("%s %s %s", s, op, f.Value()))
		if err != nil {
			return err
		}
	} else if s = f.Name(); unspecified && s != "" {
		_, err = w.WriteString(fmt.Sprintf("@%s", s))
		if err != nil {
			return err
		}
	}

	inner := f.SubFilters()
	for i := range inner {
		if i != 0 {
			_, err = w.WriteString(" " + op.String() + " ")
			if err != nil {
				return err
			}
		}

		err = writeFilterStringTo(w, inner[i])
		if err != nil {
			return err
		}
	}

	if s = f.Name(); s != "" && !unspecified {
		_, err = w.WriteString(" AS " + s)
		if err != nil {
			return err
		}
	}

	return nil
}

// DecodeString decodes PlacementPolicy from the string composed using
// WriteStringTo. Returns error if s is malformed.
func (p *PlacementPolicy) DecodeString(s string) error {
	var v policyVisitor

	input := antlr.NewInputStream(s)
	lexer := parser.NewQueryLexer(input)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(&v)
	stream := antlr.NewCommonTokenStream(lexer, 0)

	pp := parser.NewQuery(stream)
	pp.BuildParseTrees = true

	pp.RemoveErrorListeners()
	pp.AddErrorListener(&v)
	pl := pp.Policy().Accept(&v)

	if len(v.errors) != 0 {
		return v.errors[0]
	}

	parsed, ok := pl.(*PlacementPolicy)
	if !ok {
		return fmt.Errorf("unexpected parsed instance type %T", pl)
	} else if parsed == nil {
		return errors.New("parsed nil value")
	}

	if err := validatePolicy(*p); err != nil {
		return fmt.Errorf("invalid policy: %w", err)
	}

	*p = *parsed

	return nil
}

var (
	// errUnknownFilter is returned when a value of FROM in a query is unknown.
	errUnknownFilter = errors.New("filter not found")
	// errUnknownSelector is returned when a value of IN is unknown.
	errUnknownSelector = errors.New("policy: selector not found")
	// errSyntaxError is returned for errors found by ANTLR parser.
	errSyntaxError = errors.New("policy: syntax error")
)

type policyVisitor struct {
	errors []error
	parser.BaseQueryVisitor
	antlr.DefaultErrorListener
}

func (p *policyVisitor) SyntaxError(_ antlr.Recognizer, _ any, line, column int, msg string, _ antlr.RecognitionException) {
	p.reportError(fmt.Errorf("%w: line %d:%d %s", errSyntaxError, line, column, msg))
}

func (p *policyVisitor) reportError(err error) any {
	p.errors = append(p.errors, err)
	return nil
}

// VisitPolicy implements parser.QueryVisitor interface.
func (p *policyVisitor) VisitPolicy(ctx *parser.PolicyContext) any {
	if len(p.errors) != 0 {
		return nil
	}

	pl := new(PlacementPolicy)
	repStmts := ctx.AllRepStmt()
	pl.replicas = make([]ReplicaDescriptor, len(repStmts))

	for i, r := range repStmts {
		res, ok := r.Accept(p).(*protonetmap.Replica)
		if !ok {
			return nil
		}
		pl.replicas[i].fromProtoMessage(res)
	}

	if cbfStmt := ctx.CbfStmt(); cbfStmt != nil {
		cbf, ok := cbfStmt.(*parser.CbfStmtContext).Accept(p).(uint32)
		if !ok {
			return nil
		}
		pl.SetContainerBackupFactor(cbf)
	}

	selStmts := ctx.AllSelectStmt()
	pl.selectors = make([]Selector, len(selStmts))

	for i, s := range selStmts {
		res, ok := s.Accept(p).(*protonetmap.Selector)
		if !ok {
			return nil
		}
		if err := pl.selectors[i].fromProtoMessage(res); err != nil {
			return fmt.Errorf("invalid selector #%d: %w", i, err)
		}
	}

	filtStmts := ctx.AllFilterStmt()
	pl.filters = make([]Filter, len(filtStmts))

	for i, f := range filtStmts {
		res, ok := f.Accept(p).(*protonetmap.Filter)
		if !ok {
			return nil
		}
		if err := pl.filters[i].fromProtoMessage(res); err != nil {
			return fmt.Errorf("invalid filter #%d: %w", i, err)
		}
	}

	return pl
}

func (p *policyVisitor) VisitCbfStmt(ctx *parser.CbfStmtContext) any {
	cbf, err := strconv.ParseUint(ctx.GetBackupFactor().GetText(), 10, 32)
	if err != nil {
		return p.reportError(errInvalidNumber)
	}

	return uint32(cbf)
}

// VisitRepStmt implements parser.QueryVisitor interface.
func (p *policyVisitor) VisitRepStmt(ctx *parser.RepStmtContext) any {
	num, err := strconv.ParseUint(ctx.GetCount().GetText(), 10, 32)
	if err != nil {
		return p.reportError(errInvalidNumber)
	}

	rs := new(protonetmap.Replica)
	rs.Count = uint32(num)

	if sel := ctx.GetSelector(); sel != nil {
		rs.Selector = sel.GetText()
	}

	return rs
}

// VisitSelectStmt implements parser.QueryVisitor interface.
func (p *policyVisitor) VisitSelectStmt(ctx *parser.SelectStmtContext) any {
	res, err := strconv.ParseUint(ctx.GetCount().GetText(), 10, 32)
	if err != nil {
		return p.reportError(errInvalidNumber)
	}

	s := new(protonetmap.Selector)
	s.Count = uint32(res)

	if clStmt := ctx.Clause(); clStmt != nil {
		s.Clause = clauseFromString(clStmt.GetText())
	}

	if bStmt := ctx.GetBucket(); bStmt != nil {
		s.Attribute = ctx.GetBucket().GetText()
	}

	s.Filter = ctx.GetFilter().GetText() // either ident or wildcard

	if ctx.AS() != nil {
		s.Name = ctx.GetName().GetText()
	}
	return s
}

// VisitFilterStmt implements parser.QueryVisitor interface.
func (p *policyVisitor) VisitFilterStmt(ctx *parser.FilterStmtContext) any {
	f := p.VisitFilterExpr(ctx.GetExpr().(*parser.FilterExprContext)).(*protonetmap.Filter)
	f.Name = ctx.GetName().GetText()
	return f
}

func (p *policyVisitor) VisitFilterExpr(ctx *parser.FilterExprContext) any {
	if eCtx := ctx.Expr(); eCtx != nil {
		return eCtx.Accept(p)
	}

	if inner := ctx.GetInner(); inner != nil {
		return inner.Accept(p)
	}

	f := new(protonetmap.Filter)
	op := operationFromString(ctx.GetOp().GetText())
	f.Op = op

	f1 := ctx.GetF1().Accept(p).(*protonetmap.Filter)
	f2 := ctx.GetF2().Accept(p).(*protonetmap.Filter)

	// Consider f1=(.. AND ..) AND f2. This can be merged because our AND operation
	// is of arbitrary arity. ANTLR generates left-associative parse-tree by default.
	if f1.GetOp() == op {
		f.Filters = append(f1.GetFilters(), f2)
		return f
	}

	f.Filters = []*protonetmap.Filter{f1, f2}

	return f
}

// VisitFilterKey implements parser.QueryVisitor interface.
func (p *policyVisitor) VisitFilterKey(ctx *parser.FilterKeyContext) any {
	if id := ctx.Ident(); id != nil {
		return id.GetText()
	}

	str := ctx.STRING().GetText()
	return str[1 : len(str)-1]
}

func (p *policyVisitor) VisitFilterValue(ctx *parser.FilterValueContext) any {
	if id := ctx.Ident(); id != nil {
		return id.GetText()
	}

	if num := ctx.Number(); num != nil {
		return num.GetText()
	}

	str := ctx.STRING().GetText()
	return str[1 : len(str)-1]
}

// VisitExpr implements parser.QueryVisitor interface.
func (p *policyVisitor) VisitExpr(ctx *parser.ExprContext) any {
	f := new(protonetmap.Filter)
	if flt := ctx.GetFilter(); flt != nil {
		f.Name = flt.GetText()
		return f
	}

	key := ctx.GetKey().Accept(p)
	opStr := ctx.SIMPLE_OP().GetText()
	value := ctx.GetValue().Accept(p)

	f.Key = key.(string)
	f.Op = operationFromString(opStr)
	f.Value = value.(string)

	return f
}

// validatePolicy checks high-level constraints such as filter link in SELECT
// being actually defined in FILTER section.
func validatePolicy(p PlacementPolicy) error {
	seenFilters := map[string]bool{}

	for i := range p.filters {
		seenFilters[p.filters[i].Name()] = true
	}

	seenSelectors := map[string]bool{}

	for i := range p.selectors {
		if flt := p.selectors[i].FilterName(); flt != mainFilterName && !seenFilters[flt] {
			return fmt.Errorf("%w: '%s'", errUnknownFilter, flt)
		}

		seenSelectors[p.selectors[i].Name()] = true
	}

	for i := range p.replicas {
		if sel := p.replicas[i].SelectorName(); sel != "" && !seenSelectors[sel] {
			return fmt.Errorf("%w: '%s'", errUnknownSelector, sel)
		}
	}

	return nil
}

func clauseFromString(s string) protonetmap.Clause {
	switch s {
	default:
		// Such errors should be handled by ANTLR code thus this panic.
		panic(fmt.Errorf("BUG: invalid clause: %s", s))
	case "CLAUSE_UNSPECIFIED":
		return protonetmap.Clause_CLAUSE_UNSPECIFIED
	case "SAME":
		return protonetmap.Clause_SAME
	case "DISTINCT":
		return protonetmap.Clause_DISTINCT
	}
}

func operationFromString(s string) protonetmap.Operation {
	switch s {
	default:
		// Such errors should be handled by ANTLR code thus this panic.
		panic(fmt.Errorf("BUG: invalid operation: %s", s))
	case "OPERATION_UNSPECIFIED":
		return protonetmap.Operation_OPERATION_UNSPECIFIED
	case "EQ":
		return protonetmap.Operation_EQ
	case "NE":
		return protonetmap.Operation_NE
	case "GT":
		return protonetmap.Operation_GT
	case "GE":
		return protonetmap.Operation_GE
	case "LT":
		return protonetmap.Operation_LT
	case "LE":
		return protonetmap.Operation_LE
	case "OR":
		return protonetmap.Operation_OR
	case "AND":
		return protonetmap.Operation_AND
	}
}

var errInvalidNodeSetDesc = errors.New("invalid node set descriptor")

// Verify checks whether p complies with NeoFS protocol requirements. The checks
// performed may vary, so the method is recommended for system purposes only.
func (p PlacementPolicy) Verify() error {
	rs := p.Replicas()
	if len(rs) > maxContainerNodeSets {
		return fmt.Errorf("more than %d node sets", maxContainerNodeSets)
	}
	ss := p.Selectors()
	bf := p.ContainerBackupFactor()
	if bf == 0 {
		bf = defaultContainerBackupFactor
	}
	var cnrNodeCount uint32
	for i := range rs {
		rNum := rs[i].NumberOfObjects()
		if rNum > maxObjectReplicasPerSet {
			return fmt.Errorf("%w #%d: more than %d object replicas", errInvalidNodeSetDesc, i, maxObjectReplicasPerSet)
		}
		var sNum uint32
		if sName := rs[i].SelectorName(); sName != "" {
			si := slices.IndexFunc(ss, func(s Selector) bool { return s.Name() == sName })
			if si < 0 {
				return fmt.Errorf("%w #%d: missing selector %q", errInvalidNodeSetDesc, i, sName)
			}
			sNum = ss[si].NumberOfNodes()
		} else {
			sNum = rNum
		}
		nodesInSet := bf * sNum
		if nodesInSet > maxContainerNodesInSet {
			return fmt.Errorf("%w #%d: more than %d nodes", errInvalidNodeSetDesc, i, maxContainerNodesInSet)
		}
		if cnrNodeCount += nodesInSet; cnrNodeCount > maxContainerNodes {
			return fmt.Errorf("more than %d nodes in total", maxContainerNodes)
		}
	}
	return nil
}
