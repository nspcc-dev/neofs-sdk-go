package netmap

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/nspcc-dev/neofs-api-go/v2/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/netmap/parser"
)

// PlacementPolicy declares policy to store objects in the NeoFS container.
// Within itself, PlacementPolicy represents a set of rules to select a subset
// of nodes from NeoFS network map - node-candidates for object storage.
//
// PlacementPolicy is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/netmap.PlacementPolicy
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration.
type PlacementPolicy struct {
	backupFactor uint32

	filters []netmap.Filter

	selectors []netmap.Selector

	replicas []netmap.Replica
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

func copyFilter(f netmap.Filter) netmap.Filter {
	var filter netmap.Filter

	filter.SetName(f.GetName())
	filter.SetKey(f.GetKey())
	filter.SetOp(f.GetOp())
	filter.SetValue(f.GetValue())

	if f.GetFilters() != nil {
		filters := make([]netmap.Filter, len(f.GetFilters()))

		for i, internalFilter := range f.GetFilters() {
			filters[i] = copyFilter(internalFilter)
		}

		filter.SetFilters(filters)
	} else {
		filter.SetFilters(nil)
	}

	return filter
}

// CopyTo writes deep copy of the [PlacementPolicy] to dst.
func (p PlacementPolicy) CopyTo(dst *PlacementPolicy) {
	dst.SetContainerBackupFactor(p.backupFactor)

	dst.filters = make([]netmap.Filter, len(p.filters))
	for i, f := range p.filters {
		dst.filters[i] = copyFilter(f)
	}

	// netmap.Selector is a struct with simple types, no links inside. Just create a new slice and copy all items inside.
	dst.selectors = slices.Clone(p.selectors)

	// netmap.Replica is a struct with simple types, no links inside. Just create a new slice and copy all items inside.
	dst.replicas = slices.Clone(p.replicas)
}

func (p *PlacementPolicy) readFromV2(m netmap.PlacementPolicy, checkFieldPresence bool) error {
	p.replicas = m.GetReplicas()
	if checkFieldPresence && len(p.replicas) == 0 {
		return errors.New("missing replicas")
	}

	p.backupFactor = m.GetContainerBackupFactor()
	p.selectors = m.GetSelectors()
	p.filters = m.GetFilters()

	return nil
}

// Marshal encodes PlacementPolicy into a binary format of the NeoFS API
// protocol (Protocol Buffers with direct field order).
//
// See also Unmarshal.
func (p PlacementPolicy) Marshal() []byte {
	var m netmap.PlacementPolicy
	p.WriteToV2(&m)

	return m.StableMarshal(nil)
}

// Unmarshal decodes NeoFS API protocol binary format into the PlacementPolicy
// (Protocol Buffers with direct field order). Returns an error describing
// a format violation.
//
// See also Marshal.
func (p *PlacementPolicy) Unmarshal(data []byte) error {
	var m netmap.PlacementPolicy

	err := m.Unmarshal(data)
	if err != nil {
		return err
	}

	return p.readFromV2(m, false)
}

// MarshalJSON encodes PlacementPolicy into a JSON format of the NeoFS API
// protocol (Protocol Buffers JSON).
//
// See also UnmarshalJSON.
func (p PlacementPolicy) MarshalJSON() ([]byte, error) {
	var m netmap.PlacementPolicy
	p.WriteToV2(&m)

	return m.MarshalJSON()
}

// UnmarshalJSON decodes NeoFS API protocol JSON format into the PlacementPolicy
// (Protocol Buffers JSON). Returns an error describing a format violation.
//
// See also MarshalJSON.
func (p *PlacementPolicy) UnmarshalJSON(data []byte) error {
	var m netmap.PlacementPolicy

	err := m.UnmarshalJSON(data)
	if err != nil {
		return err
	}

	return p.readFromV2(m, false)
}

// ReadFromV2 reads PlacementPolicy from the netmap.PlacementPolicy message.
// Checks if the message conforms to NeoFS API V2 protocol.
//
// See also WriteToV2.
func (p *PlacementPolicy) ReadFromV2(m netmap.PlacementPolicy) error {
	return p.readFromV2(m, true)
}

// WriteToV2 writes PlacementPolicy to the session.Token message.
// The message must not be nil.
//
// See also ReadFromV2.
func (p PlacementPolicy) WriteToV2(m *netmap.PlacementPolicy) {
	m.SetContainerBackupFactor(p.backupFactor)
	m.SetFilters(p.filters)
	m.SetSelectors(p.selectors)
	m.SetReplicas(p.replicas)
}

// ReplicaDescriptor replica descriptor characterizes replicas of objects from
// the subset selected by a particular Selector.
type ReplicaDescriptor struct {
	m netmap.Replica
}

// SetNumberOfObjects sets number of object replicas.
func (r *ReplicaDescriptor) SetNumberOfObjects(c uint32) {
	r.m.SetCount(c)
}

// NumberOfObjects returns number set using SetNumberOfObjects.
//
// Zero ReplicaDescriptor has zero number of objects.
func (r ReplicaDescriptor) NumberOfObjects() uint32 {
	return r.m.GetCount()
}

// SetSelectorName sets name of the related Selector.
//
// Zero ReplicaDescriptor references to the root bucket's selector: it contains
// all possible nodes to store the object.
//
// See also [ReplicaDescriptor.SelectorName].
func (r *ReplicaDescriptor) SetSelectorName(s string) {
	r.m.SetSelector(s)
}

// SelectorName returns name of the related Selector.
//
// Zero ReplicaDescriptor references to the root bucket's selector: it contains
// all possible nodes to store the object.
//
// See also [ReplicaDescriptor.SetSelectorName].
func (r ReplicaDescriptor) SelectorName() string {
	return r.m.GetSelector()
}

// SetReplicas sets list of object replica's characteristics.
//
// See also [PlacementPolicy.Replicas], [PlacementPolicy.NumberOfReplicas],
// [PlacementPolicy.ReplicaNumberByIndex].
func (p *PlacementPolicy) SetReplicas(rs []ReplicaDescriptor) {
	p.replicas = make([]netmap.Replica, len(rs))

	for i := range rs {
		p.replicas[i] = rs[i].m
	}
}

// Replicas returns list of object replica characteristics.
//
// See also [PlacementPolicy.SetReplicas], [PlacementPolicy.NumberOfReplicas],
// [PlacementPolicy.ReplicaNumberByIndex].
func (p PlacementPolicy) Replicas() []ReplicaDescriptor {
	rs := make([]ReplicaDescriptor, len(p.replicas))
	for i := range p.replicas {
		rs[i].m = p.replicas[i]
	}
	return rs
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
	return p.replicas[i].GetCount()
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
	m netmap.Selector
}

// SetName sets name with which the Selector can be referenced.
//
// Zero Selector is unnamed.
func (s *Selector) SetName(name string) {
	s.m.SetName(name)
}

// Name returns name with which the Selector can be referenced.
//
// Zero Selector is unnamed.
//
// See also [Selector.Name].
func (s Selector) Name() string {
	return s.m.GetName()
}

// SetNumberOfNodes sets number of nodes to select from the bucket.
//
// Zero Selector selects nothing.
//
// See also [Selector.NumberOfNodes].
func (s *Selector) SetNumberOfNodes(num uint32) {
	s.m.SetCount(num)
}

// NumberOfNodes returns number of nodes to select from the bucket.
//
// Zero Selector selects nothing.
//
// See also [Selector.SetNumberOfNodes].
func (s Selector) NumberOfNodes() uint32 {
	return s.m.GetCount()
}

// SelectByBucketAttribute sets attribute of the bucket to select nodes from.
//
// Zero Selector has empty attribute.
//
// See also [Selector.BucketAttribute].
func (s *Selector) SelectByBucketAttribute(bucket string) {
	s.m.SetAttribute(bucket)
}

// BucketAttribute returns attribute of the bucket to select nodes from.
//
// Zero Selector has empty attribute.
//
// See also [Selector.SelectByBucketAttribute].
func (s *Selector) BucketAttribute() string {
	return s.m.GetAttribute()
}

// SelectSame makes selection algorithm to select only nodes having the same values
// of the bucket attribute.
//
// Zero Selector doesn't specify selection modifier so nodes are selected randomly.
//
// See also [Selector.SelectByBucketAttribute], [Selector.IsSame].
func (s *Selector) SelectSame() {
	s.m.SetClause(netmap.Same)
}

// IsSame checks whether selection algorithm is set to select only nodes having
// the same values of the bucket attribute.
//
// See also [Selector.SelectSame].
func (s *Selector) IsSame() bool {
	return s.m.GetClause() == netmap.Same
}

// SelectDistinct makes selection algorithm to select only nodes having the different values
// of the bucket attribute.
//
// Zero Selector doesn't specify selection modifier so nodes are selected randomly.
//
// See also [Selector.SelectByBucketAttribute], [Selector.IsDistinct].
func (s *Selector) SelectDistinct() {
	s.m.SetClause(netmap.Distinct)
}

// IsDistinct checks whether selection algorithm is set to select only nodes
// having the different values of the bucket attribute.
//
// See also [Selector.SelectByBucketAttribute], [Selector.SelectDistinct].
func (s *Selector) IsDistinct() bool {
	return s.m.GetClause() == netmap.Distinct
}

// SetFilterName sets reference to pre-filtering nodes for selection.
//
// Zero Selector has no filtering reference.
//
// See also Filter.SetName.
func (s *Selector) SetFilterName(f string) {
	s.m.SetFilter(f)
}

// FilterName returns reference to pre-filtering nodes for selection.
//
// Zero Selector has no filtering reference.
//
// See also [Filter.SetName], [Selector.SetFilterName].
func (s *Selector) FilterName() string {
	return s.m.GetFilter()
}

// SetSelectors sets list of Selector to form the subset of the nodes to store
// container objects.
//
// Zero PlacementPolicy does not declare selectors.
//
// See also [PlacementPolicy.Selectors].
func (p *PlacementPolicy) SetSelectors(ss []Selector) {
	p.selectors = make([]netmap.Selector, len(ss))

	for i := range ss {
		p.selectors[i] = ss[i].m
	}
}

// Selectors returns list of Selector to form the subset of the nodes to store
// container objects.
//
// Zero PlacementPolicy does not declare selectors.
//
// See also [PlacementPolicy.SetSelectors].
func (p PlacementPolicy) Selectors() []Selector {
	ss := make([]Selector, len(p.selectors))
	for i := range p.selectors {
		ss[i].m = p.selectors[i]
	}
	return ss
}

// Filter contains rules for filtering the node sets.
type Filter struct {
	m netmap.Filter
}

// SetName sets name with which the Filter can be referenced or, for inner filters,
// to which the Filter references. Top-level filters MUST be named. The name
// MUST NOT be '*'.
//
// Zero Filter is unnamed.
//
// See also [Filter.Name].
func (x *Filter) SetName(name string) {
	x.m.SetName(name)
}

// Name returns name with which the Filter can be referenced or, for inner
// filters, to which the Filter references. Top-level filters MUST be named. The
// name MUST NOT be '*'.
//
// Zero Filter is unnamed.
//
// See also [Filter.SetName].
func (x Filter) Name() string {
	return x.m.GetName()
}

// Key returns key to the property.
func (x Filter) Key() string {
	return x.m.GetKey()
}

// Op returns operator to match the property.
func (x Filter) Op() FilterOp {
	return FilterOp(x.m.GetOp())
}

// Value returns value to check the property against.
func (x Filter) Value() string {
	return x.m.GetValue()
}

// SubFilters returns list of sub-filters when Filter is complex.
func (x Filter) SubFilters() []Filter {
	fsm := x.m.GetFilters()
	if len(fsm) == 0 {
		return nil
	}

	fs := make([]Filter, len(fsm))
	for i := range fsm {
		fs[i] = Filter{m: fsm[i]}
	}

	return fs
}

func (x *Filter) setAttribute(key string, op netmap.Operation, val string) {
	x.m.SetKey(key)
	x.m.SetOp(op)
	x.m.SetValue(val)
}

// Equal applies the rule to accept only nodes with the same attribute value.
//
// Method SHOULD NOT be called along with other similar methods.
func (x *Filter) Equal(key, value string) {
	x.setAttribute(key, netmap.EQ, value)
}

// NotEqual applies the rule to accept only nodes with the distinct attribute value.
//
// Method SHOULD NOT be called along with other similar methods.
func (x *Filter) NotEqual(key, value string) {
	x.setAttribute(key, netmap.NE, value)
}

// NumericGT applies the rule to accept only nodes with the numeric attribute
// greater than given number.
//
// Method SHOULD NOT be called along with other similar methods.
func (x *Filter) NumericGT(key string, num int64) {
	x.setAttribute(key, netmap.GT, strconv.FormatInt(num, 10))
}

// NumericGE applies the rule to accept only nodes with the numeric attribute
// greater than or equal to given number.
//
// Method SHOULD NOT be called along with other similar methods.
func (x *Filter) NumericGE(key string, num int64) {
	x.setAttribute(key, netmap.GE, strconv.FormatInt(num, 10))
}

// NumericLT applies the rule to accept only nodes with the numeric attribute
// less than given number.
//
// Method SHOULD NOT be called along with other similar methods.
func (x *Filter) NumericLT(key string, num int64) {
	x.setAttribute(key, netmap.LT, strconv.FormatInt(num, 10))
}

// NumericLE applies the rule to accept only nodes with the numeric attribute
// less than or equal to given number.
//
// Method SHOULD NOT be called along with other similar methods.
func (x *Filter) NumericLE(key string, num int64) {
	x.setAttribute(key, netmap.LE, strconv.FormatInt(num, 10))
}

func (x *Filter) setInnerFilters(op netmap.Operation, filters []Filter) {
	x.setAttribute("", op, "")

	inner := x.m.GetFilters()
	if rem := len(filters) - len(inner); rem > 0 {
		inner = append(inner, make([]netmap.Filter, rem)...)
	}

	for i := range filters {
		inner[i] = filters[i].m
	}

	x.m.SetFilters(inner)
}

// LogicalOR applies the rule to accept only nodes which satisfy at least one
// of the given filters.
//
// Method SHOULD NOT be called along with other similar methods.
func (x *Filter) LogicalOR(filters ...Filter) {
	x.setInnerFilters(netmap.OR, filters)
}

// LogicalAND applies the rule to accept only nodes which satisfy all the given
// filters.
//
// Method SHOULD NOT be called along with other similar methods.
func (x *Filter) LogicalAND(filters ...Filter) {
	x.setInnerFilters(netmap.AND, filters)
}

// Filters returns list of Filter that will be applied when selecting nodes.
//
// Zero PlacementPolicy has no filters.
//
// See also [PlacementPolicy.SetFilters].
func (p PlacementPolicy) Filters() []Filter {
	fs := make([]Filter, len(p.filters))
	for i := range p.filters {
		fs[i] = Filter{m: p.filters[i]}
	}
	return fs
}

// SetFilters sets list of Filter that will be applied when selecting nodes.
//
// Zero PlacementPolicy has no filters.
//
// See also [PlacementPolicy.Filters].
func (p *PlacementPolicy) SetFilters(fs []Filter) {
	p.filters = make([]netmap.Filter, len(fs))

	for i := range fs {
		p.filters[i] = fs[i].m
	}
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

		c := p.replicas[i].GetCount()
		s := p.replicas[i].GetSelector()

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

		_, err = w.WriteString(fmt.Sprintf("SELECT %d", p.selectors[i].GetCount()))
		if err != nil {
			return err
		}

		if s = p.selectors[i].GetAttribute(); s != "" {
			var clause string

			switch p.selectors[i].GetClause() {
			case netmap.Same:
				clause = "SAME "
			case netmap.Distinct:
				clause = "DISTINCT "
			default:
				clause = ""
			}

			_, err = w.WriteString(fmt.Sprintf(" IN %s%s", clause, s))
			if err != nil {
				return err
			}
		}

		if s = p.selectors[i].GetFilter(); s != "" {
			_, err = w.WriteString(" FROM " + s)
			if err != nil {
				return err
			}
		}

		if s = p.selectors[i].GetName(); s != "" {
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

func writeFilterStringTo(w io.StringWriter, f netmap.Filter) error {
	var err error
	var s string
	op := f.GetOp()
	unspecified := op == 0

	if s = f.GetKey(); s != "" {
		_, err = w.WriteString(fmt.Sprintf("%s %s %s", s, op, f.GetValue()))
		if err != nil {
			return err
		}
	} else if s = f.GetName(); unspecified && s != "" {
		_, err = w.WriteString(fmt.Sprintf("@%s", s))
		if err != nil {
			return err
		}
	}

	inner := f.GetFilters()
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

	if s = f.GetName(); s != "" && !unspecified {
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
	pl.replicas = make([]netmap.Replica, 0, len(repStmts))

	for _, r := range repStmts {
		res, ok := r.Accept(p).(*netmap.Replica)
		if !ok {
			return nil
		}

		pl.replicas = append(pl.replicas, *res)
	}

	if cbfStmt := ctx.CbfStmt(); cbfStmt != nil {
		cbf, ok := cbfStmt.(*parser.CbfStmtContext).Accept(p).(uint32)
		if !ok {
			return nil
		}
		pl.SetContainerBackupFactor(cbf)
	}

	selStmts := ctx.AllSelectStmt()
	pl.selectors = make([]netmap.Selector, 0, len(selStmts))

	for _, s := range selStmts {
		res, ok := s.Accept(p).(*netmap.Selector)
		if !ok {
			return nil
		}

		pl.selectors = append(pl.selectors, *res)
	}

	filtStmts := ctx.AllFilterStmt()
	pl.filters = make([]netmap.Filter, 0, len(filtStmts))

	for _, f := range filtStmts {
		pl.filters = append(pl.filters, *f.Accept(p).(*netmap.Filter))
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
	rs.SetCount(uint32(num))

	if sel := ctx.GetSelector(); sel != nil {
		rs.SetSelector(sel.GetText())
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
	s.SetCount(uint32(res))

	if clStmt := ctx.Clause(); clStmt != nil {
		s.SetClause(clauseFromString(clStmt.GetText()))
	}

	if bStmt := ctx.GetBucket(); bStmt != nil {
		s.SetAttribute(ctx.GetBucket().GetText())
	}

	s.SetFilter(ctx.GetFilter().GetText()) // either ident or wildcard

	if ctx.AS() != nil {
		s.SetName(ctx.GetName().GetText())
	}
	return s
}

// VisitFilterStmt implements parser.QueryVisitor interface.
func (p *policyVisitor) VisitFilterStmt(ctx *parser.FilterStmtContext) any {
	f := p.VisitFilterExpr(ctx.GetExpr().(*parser.FilterExprContext)).(*netmap.Filter)
	f.SetName(ctx.GetName().GetText())
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
	f.SetOp(op)

	f1 := *ctx.GetF1().Accept(p).(*netmap.Filter)
	f2 := *ctx.GetF2().Accept(p).(*netmap.Filter)

	// Consider f1=(.. AND ..) AND f2. This can be merged because our AND operation
	// is of arbitrary arity. ANTLR generates left-associative parse-tree by default.
	if f1.GetOp() == op {
		f.SetFilters(append(f1.GetFilters(), f2))
		return f
	}

	f.SetFilters([]netmap.Filter{f1, f2})

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
		f.SetName(flt.GetText())
		return f
	}

	key := ctx.GetKey().Accept(p)
	opStr := ctx.SIMPLE_OP().GetText()
	value := ctx.GetValue().Accept(p)

	f.SetKey(key.(string))
	f.SetOp(operationFromString(opStr))
	f.SetValue(value.(string))

	return f
}

// validatePolicy checks high-level constraints such as filter link in SELECT
// being actually defined in FILTER section.
func validatePolicy(p PlacementPolicy) error {
	seenFilters := map[string]bool{}

	for i := range p.filters {
		seenFilters[p.filters[i].GetName()] = true
	}

	seenSelectors := map[string]bool{}

	for i := range p.selectors {
		if flt := p.selectors[i].GetFilter(); flt != mainFilterName && !seenFilters[flt] {
			return fmt.Errorf("%w: '%s'", errUnknownFilter, flt)
		}

		seenSelectors[p.selectors[i].GetName()] = true
	}

	for i := range p.replicas {
		if sel := p.replicas[i].GetSelector(); sel != "" && !seenSelectors[sel] {
			return fmt.Errorf("%w: '%s'", errUnknownSelector, sel)
		}
	}

	return nil
}

func clauseFromString(s string) (c netmap.Clause) {
	if !c.FromString(strings.ToUpper(s)) {
		// Such errors should be handled by ANTLR code thus this panic.
		panic(fmt.Errorf("BUG: invalid clause: %s", c))
	}

	return
}

func operationFromString(s string) (op netmap.Operation) {
	if !op.FromString(strings.ToUpper(s)) {
		// Such errors should be handled by ANTLR code thus this panic.
		panic(fmt.Errorf("BUG: invalid operation: %s", op))
	}

	return
}
