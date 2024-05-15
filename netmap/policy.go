package netmap

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/nspcc-dev/neofs-sdk-go/api/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/netmap/parser"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// PlacementPolicy declares policy to store objects in the NeoFS container.
// Within itself, PlacementPolicy represents a set of rules to select a subset
// of nodes from NeoFS network map - node-candidates for object storage.
//
// PlacementPolicy is mutually compatible with [netmap.PlacementPolicy] message.
// See [PlacementPolicy.ReadFromV2] / [PlacementPolicy.WriteToV2] methods.
//
// Instances can be created using built-in var declaration.
type PlacementPolicy struct {
	backupFactor uint32

	filters []Filter

	selectors []Selector

	replicas []ReplicaDescriptor
}

// FilterOp defines the matching property.
type FilterOp uint32

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
	res := f

	if f.subs != nil {
		res.subs = make([]Filter, len(f.subs))
		for i := range f.subs {
			res.subs[i] = copyFilter(f.subs[i])
		}
	} else {
		res.subs = nil
	}

	return res
}

// CopyTo writes deep copy of the [PlacementPolicy] to dst.
func (p PlacementPolicy) CopyTo(dst *PlacementPolicy) {
	dst.SetContainerBackupFactor(p.backupFactor)

	if p.filters != nil {
		dst.filters = make([]Filter, len(p.filters))
		for i, f := range p.filters {
			dst.filters[i] = copyFilter(f)
		}
	} else {
		dst.filters = nil
	}

	if p.selectors != nil {
		dst.selectors = make([]Selector, len(p.selectors))
		copy(dst.selectors, p.selectors)
	} else {
		dst.selectors = nil
	}

	if p.replicas != nil {
		dst.replicas = make([]ReplicaDescriptor, len(p.replicas))
		copy(dst.replicas, p.replicas)
	} else {
		dst.replicas = nil
	}
}

func (p *PlacementPolicy) readFromV2(m *netmap.PlacementPolicy, checkFieldPresence bool) error {
	if checkFieldPresence && len(m.Replicas) == 0 {
		return errors.New("missing replicas")
	}

	if m.Replicas != nil {
		p.replicas = make([]ReplicaDescriptor, len(m.Replicas))
		for i := range m.Replicas {
			p.replicas[i] = replicaFromAPI(m.Replicas[i])
		}
	} else {
		p.replicas = nil
	}

	if m.Selectors != nil {
		p.selectors = make([]Selector, len(m.Selectors))
		for i := range m.Selectors {
			p.selectors[i] = selectorFromAPI(m.Selectors[i])
		}
	} else {
		p.selectors = nil
	}

	p.filters = filtersFromAPI(m.Filters)
	p.backupFactor = m.ContainerBackupFactor

	return nil
}

// Marshal encodes PlacementPolicy into a binary format of the NeoFS API
// protocol (Protocol Buffers V3 with direct field order).
//
// See also [PlacementPolicy.Unmarshal].
func (p PlacementPolicy) Marshal() []byte {
	var m netmap.PlacementPolicy
	p.WriteToV2(&m)
	b := make([]byte, m.MarshaledSize())
	m.MarshalStable(b)
	return b
}

// Unmarshal decodes Protocol Buffers V3 binary data into the PlacementPolicy.
// Returns an error describing a format violation of the specified fields.
// Unmarshal does not check presence of the required fields and, at the same
// time, checks format of presented fields.
//
// See also [PlacementPolicy.Marshal].
func (p *PlacementPolicy) Unmarshal(data []byte) error {
	var m netmap.PlacementPolicy
	err := proto.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protobuf: %w", err)
	}

	return p.readFromV2(&m, false)
}

// MarshalJSON encodes PlacementPolicy into a JSON format of the NeoFS API
// protocol (Protocol Buffers V3 JSON).
//
// See also [Token.UnmarshalJSON].
func (p PlacementPolicy) MarshalJSON() ([]byte, error) {
	var m netmap.PlacementPolicy
	p.WriteToV2(&m)

	return protojson.Marshal(&m)
}

// UnmarshalJSON decodes NeoFS API protocol JSON data into the PlacementPolicy
// (Protocol Buffers V3 JSON). Returns an error describing a format violation.
// UnmarshalJSON does not check presence of the required fields and, at the same
// time, checks format of presented fields.
//
// See also [PlacementPolicy.MarshalJSON].
func (p *PlacementPolicy) UnmarshalJSON(data []byte) error {
	var m netmap.PlacementPolicy
	err := protojson.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protojson: %w", err)
	}

	return p.readFromV2(&m, false)
}

// ReadFromV2 reads PlacementPolicy from the [netmap.PlacementPolicy] message.
// Returns an error if the message is malformed according to the NeoFS API V2
// protocol. The message must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [PlacementPolicy.WriteToV2].
func (p *PlacementPolicy) ReadFromV2(m *netmap.PlacementPolicy) error {
	return p.readFromV2(m, true)
}

// WriteToV2 writes PlacementPolicy to the [netmap.PlacementPolicy] message of
// the NeoFS API protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [PlacementPolicy.ReadFromV2].
func (p PlacementPolicy) WriteToV2(m *netmap.PlacementPolicy) {
	if p.replicas != nil {
		m.Replicas = make([]*netmap.Replica, len(p.replicas))
		for i := range p.replicas {
			m.Replicas[i] = replicaToAPI(p.replicas[i])
		}
	} else {
		m.Replicas = nil
	}

	if p.selectors != nil {
		m.Selectors = make([]*netmap.Selector, len(p.selectors))
		for i := range p.selectors {
			m.Selectors[i] = selectorToAPI(p.selectors[i])
		}
	} else {
		m.Selectors = nil
	}

	m.Filters = filtersToAPI(p.filters)
	m.ContainerBackupFactor = p.backupFactor
}

// ReplicaDescriptor replica descriptor characterizes replicas of objects from
// the subset selected by a particular Selector.
type ReplicaDescriptor struct {
	count    uint32
	selector string
}

func replicaFromAPI(m *netmap.Replica) ReplicaDescriptor {
	var res ReplicaDescriptor
	if m != nil {
		res.count = m.Count
		res.selector = m.Selector
	}
	return res
}

func isEmptyReplica(r ReplicaDescriptor) bool {
	return r.count == 0 && r.selector == ""
}

func replicaToAPI(r ReplicaDescriptor) *netmap.Replica {
	if isEmptyReplica(r) {
		return nil
	}
	return &netmap.Replica{
		Count:    r.count,
		Selector: r.selector,
	}
}

// SetNumberOfObjects sets number of object replicas.
//
// See also [ReplicaDescriptor.NumberOfObjects].
func (r *ReplicaDescriptor) SetNumberOfObjects(c uint32) {
	r.count = c
}

// NumberOfObjects returns number set using [ReplicaDescriptor.SetNumberOfObjects].
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
//
// See also [PlacementPolicy.SetReplicas], [PlacementPolicy.NumberOfReplicas],
// [PlacementPolicy.Replicas].
func (p PlacementPolicy) NumberOfReplicas() int {
	return len(p.replicas)
}

// ReplicaNumberByIndex returns number of object replicas from the i-th replica
// descriptor. Index MUST be in range 0:[PlacementPolicy.NumberOfReplicas].
//
// Zero PlacementPolicy has no replicas.
//
// See also [PlacementPolicy.SetReplicas], [PlacementPolicy.Replicas].
func (p PlacementPolicy) ReplicaNumberByIndex(i int) uint32 {
	return p.replicas[i].count
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
	clause netmap.Clause
	attr   string
	filter string
}

func selectorFromAPI(m *netmap.Selector) Selector {
	var res Selector
	if m != nil {
		res.name = m.Name
		res.count = m.Count
		res.clause = m.Clause
		res.attr = m.Attribute
		res.filter = m.Filter
	}
	return res
}

func isEmptySelector(r Selector) bool {
	return r.count == 0 && r.name == "" && r.clause == 0 && r.filter == "" && r.attr == ""
}

func selectorToAPI(s Selector) *netmap.Selector {
	if isEmptySelector(s) {
		return nil
	}
	return &netmap.Selector{
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
//
// See also [Selector.Name].
func (s *Selector) SetName(name string) {
	s.name = name
}

// Name returns name with which the Selector can be referenced.
//
// Zero Selector is unnamed.
//
// See also [Selector.SetName].
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
	s.clause = netmap.Clause_SAME
}

// IsSame checks whether selection algorithm is set to select only nodes having
// the same values of the bucket attribute.
//
// See also [Selector.SelectSame].
func (s *Selector) IsSame() bool {
	return s.clause == netmap.Clause_SAME
}

// SelectDistinct makes selection algorithm to select only nodes having the different values
// of the bucket attribute.
//
// Zero Selector doesn't specify selection modifier so nodes are selected randomly.
//
// See also [Selector.SelectByBucketAttribute], [Selector.IsDistinct].
func (s *Selector) SelectDistinct() {
	s.clause = netmap.Clause_DISTINCT
}

// IsDistinct checks whether selection algorithm is set to select only nodes
// having the different values of the bucket attribute.
//
// See also [Selector.SelectByBucketAttribute], [Selector.SelectDistinct].
func (s *Selector) IsDistinct() bool {
	return s.clause == netmap.Clause_DISTINCT
}

// SetFilterName sets reference to pre-filtering nodes for selection.
//
// Zero Selector has no filtering reference.
//
// See also [Selector.FilterName].
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

func filterFromAPI(f *netmap.Filter) Filter {
	var res Filter
	if f != nil {
		res.name = f.Name
		res.key = f.Key
		res.op = FilterOp(f.Op)
		res.val = f.Value
		if len(f.Filters) > 0 {
			res.subs = filtersFromAPI(f.Filters)
		}
	}
	return res
}

func filtersFromAPI(fs []*netmap.Filter) []Filter {
	if fs == nil {
		return nil
	}
	res := make([]Filter, len(fs))
	for i := range fs {
		res[i] = filterFromAPI(fs[i])
	}
	return res
}

func isEmptyFilter(f Filter) bool {
	return f.op == 0 && f.name == "" && f.key == "" && f.val == "" && len(f.subs) == 0
}

func filtersToAPI(fs []Filter) []*netmap.Filter {
	if fs == nil {
		return nil
	}
	res := make([]*netmap.Filter, len(fs))
	for i := range fs {
		if !isEmptyFilter(fs[i]) {
			res[i] = &netmap.Filter{
				Name:    fs[i].name,
				Key:     fs[i].key,
				Op:      netmap.Operation(fs[i].op),
				Value:   fs[i].val,
				Filters: filtersToAPI(fs[i].subs),
			}
		}
	}
	return res
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

		if p.replicas[i].selector != "" {
			_, err = w.WriteString(fmt.Sprintf("REP %d IN %s", p.replicas[i].count, p.replicas[i].selector))
		} else {
			_, err = w.WriteString(fmt.Sprintf("REP %d", p.replicas[i].count))
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

	for i := range p.selectors {
		err = writeLnIfNeeded()
		if err != nil {
			return err
		}

		_, err = w.WriteString(fmt.Sprintf("SELECT %d", p.selectors[i].count))
		if err != nil {
			return err
		}

		if p.selectors[i].attr != "" {
			var clause string

			switch p.selectors[i].clause {
			case netmap.Clause_SAME:
				clause = "SAME "
			case netmap.Clause_DISTINCT:
				clause = "DISTINCT "
			default:
				clause = ""
			}

			_, err = w.WriteString(fmt.Sprintf(" IN %s%s", clause, p.selectors[i].attr))
			if err != nil {
				return err
			}
		}

		if p.selectors[i].filter != "" {
			_, err = w.WriteString(" FROM " + p.selectors[i].filter)
			if err != nil {
				return err
			}
		}

		if p.selectors[i].name != "" {
			_, err = w.WriteString(" AS " + p.selectors[i].name)
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
	unspecified := f.op == 0

	if f.key != "" {
		_, err = w.WriteString(fmt.Sprintf("%s %s %s", f.key, f.op, f.val))
		if err != nil {
			return err
		}
	} else if unspecified && f.name != "" {
		_, err = w.WriteString(fmt.Sprintf("@%s", f.name))
		if err != nil {
			return err
		}
	}

	for i := range f.subs {
		if i != 0 {
			_, err = w.WriteString(" " + f.op.String() + " ")
			if err != nil {
				return err
			}
		}

		err = writeFilterStringTo(w, f.subs[i])
		if err != nil {
			return err
		}
	}

	if f.name != "" && !unspecified {
		_, err = w.WriteString(" AS " + f.name)
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
	pl.replicas = make([]ReplicaDescriptor, 0, len(repStmts))

	for _, r := range repStmts {
		res, ok := r.Accept(p).(*netmap.Replica)
		if !ok {
			return nil
		}

		pl.replicas = append(pl.replicas, replicaFromAPI(res))
	}

	if cbfStmt := ctx.CbfStmt(); cbfStmt != nil {
		cbf, ok := cbfStmt.(*parser.CbfStmtContext).Accept(p).(uint32)
		if !ok {
			return nil
		}
		pl.SetContainerBackupFactor(cbf)
	}

	selStmts := ctx.AllSelectStmt()
	pl.selectors = make([]Selector, 0, len(selStmts))

	for _, s := range selStmts {
		res, ok := s.Accept(p).(*netmap.Selector)
		if !ok {
			return nil
		}

		pl.selectors = append(pl.selectors, selectorFromAPI(res))
	}

	filtStmts := ctx.AllFilterStmt()
	pl.filters = make([]Filter, 0, len(filtStmts))

	for _, f := range filtStmts {
		pl.filters = append(pl.filters, filterFromAPI(f.Accept(p).(*netmap.Filter)))
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

	rs := new(netmap.Replica)
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

	s := new(netmap.Selector)
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
	f := p.VisitFilterExpr(ctx.GetExpr().(*parser.FilterExprContext)).(*netmap.Filter)
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

	f := new(netmap.Filter)
	op := operationFromString(ctx.GetOp().GetText())
	f.Op = op

	f1 := ctx.GetF1().Accept(p).(*netmap.Filter)
	f2 := ctx.GetF2().Accept(p).(*netmap.Filter)

	// Consider f1=(.. AND ..) AND f2. This can be merged because our AND operation
	// is of arbitrary arity. ANTLR generates left-associative parse-tree by default.
	if f1.GetOp() == op {
		f.Filters = append(f1.GetFilters(), f2)
		return f
	}

	f.Filters = []*netmap.Filter{f1, f2}

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
	f := new(netmap.Filter)
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
		seenFilters[p.filters[i].name] = true
	}

	seenSelectors := map[string]bool{}

	for i := range p.selectors {
		if p.selectors[i].filter != mainFilterName && !seenFilters[p.selectors[i].filter] {
			return fmt.Errorf("%w: '%s'", errUnknownFilter, p.selectors[i].filter)
		}

		seenSelectors[p.selectors[i].name] = true
	}

	for i := range p.replicas {
		if p.replicas[i].selector != "" && !seenSelectors[p.replicas[i].selector] {
			return fmt.Errorf("%w: '%s'", errUnknownSelector, p.replicas[i].selector)
		}
	}

	return nil
}

func clauseFromString(s string) netmap.Clause {
	v, ok := netmap.Clause_value[strings.ToUpper(s)]
	if !ok {
		// Such errors should be handled by ANTLR code thus this panic.
		panic(fmt.Errorf("BUG: invalid clause: %s", s))
	}

	return netmap.Clause(v)
}

func operationFromString(s string) netmap.Operation {
	v, ok := netmap.Operation_value[strings.ToUpper(s)]
	if !ok {
		// Such errors should be handled by ANTLR code thus this panic.
		panic(fmt.Errorf("BUG: invalid operation: %s", s))
	}

	return netmap.Operation(v)
}
