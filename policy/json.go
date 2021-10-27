package policy

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nspcc-dev/neofs-sdk-go/netmap"
)

type (
	filter struct {
		Name    string   `json:"name,omitempty"`
		Key     string   `json:"key,omitempty"`
		Op      string   `json:"op,omitempty"`
		Value   string   `json:"value,omitempty"`
		Filters []filter `json:"filters,omitempty"`
	}
	replica struct {
		Count    uint32 `json:"count"`
		Selector string `json:"selector,omitempty"`
	}
	selector struct {
		Count     uint32 `json:"count"`
		Attribute string `json:"attribute"`
		Filter    string `json:"filter,omitempty"`
		Name      string `json:"name,omitempty"`
		Clause    string `json:"clause,omitempty"`
	}
	placement struct {
		Replicas  []replica  `json:"replicas"`
		CBF       uint32     `json:"container_backup_factor,omitempty"`
		Selectors []selector `json:"selectors,omitempty"`
		Filters   []filter   `json:"filters,omitempty"`
	}
)

// ToJSON converts placement policy to JSON.
func ToJSON(np *netmap.PlacementPolicy) ([]byte, error) {
	p := new(placement)
	p.CBF = np.ContainerBackupFactor()
	p.Filters = make([]filter, len(np.Filters()))
	for i, f := range np.Filters() {
		p.Filters[i].fromNetmap(f)
	}
	p.Selectors = make([]selector, len(np.Selectors()))
	for i, s := range np.Selectors() {
		p.Selectors[i].fromNetmap(s)
	}
	p.Replicas = make([]replica, len(np.Replicas()))
	for i, r := range np.Replicas() {
		p.Replicas[i].fromNetmap(r)
	}
	return json.Marshal(p)
}

// FromJSON creates placement policy from JSON.
func FromJSON(data []byte) (*netmap.PlacementPolicy, error) {
	p := new(placement)
	if err := json.Unmarshal(data, p); err != nil {
		return nil, err
	}

	rs := make([]*netmap.Replica, len(p.Replicas))
	for i := range p.Replicas {
		rs[i] = p.Replicas[i].toNetmap()
	}

	var fs []*netmap.Filter
	if len(p.Filters) != 0 {
		fs = make([]*netmap.Filter, len(p.Filters))
		for i := range p.Filters {
			f, err := p.Filters[i].toNetmap()
			if err != nil {
				return nil, err
			}
			fs[i] = f
		}
	}

	var ss []*netmap.Selector
	if len(p.Selectors) != 0 {
		ss = make([]*netmap.Selector, len(p.Selectors))
		for i := range p.Selectors {
			s, err := p.Selectors[i].toNetmap()
			if err != nil {
				return nil, err
			}
			ss[i] = s
		}
	}

	pp := new(netmap.PlacementPolicy)
	pp.SetReplicas(rs...)
	pp.SetContainerBackupFactor(p.CBF)
	pp.SetFilters(fs...)
	pp.SetSelectors(ss...)
	return pp, nil
}

func (r *replica) toNetmap() *netmap.Replica {
	nr := new(netmap.Replica)
	nr.SetCount(r.Count)
	nr.SetSelector(r.Selector)
	return nr
}

func (r *replica) fromNetmap(nr *netmap.Replica) {
	r.Count = nr.Count()
	r.Selector = nr.Selector()
}

func (f *filter) toNetmap() (*netmap.Filter, error) {
	var op netmap.Operation
	switch strings.ToUpper(f.Op) {
	case "EQ":
		op = netmap.OpEQ
	case "NE":
		op = netmap.OpNE
	case "GT":
		op = netmap.OpGT
	case "GE":
		op = netmap.OpGE
	case "LT":
		op = netmap.OpLT
	case "LE":
		op = netmap.OpLE
	case "AND":
		op = netmap.OpAND
	case "OR":
		op = netmap.OpOR
	case "":
	default:
		return nil, fmt.Errorf("%w: '%s'", ErrUnknownOp, f.Op)
	}

	var fs []*netmap.Filter
	if len(f.Filters) != 0 {
		fs = make([]*netmap.Filter, len(f.Filters))
		for i := range f.Filters {
			var err error
			fs[i], err = f.Filters[i].toNetmap()
			if err != nil {
				return nil, err
			}
		}
	}

	nf := new(netmap.Filter)
	nf.SetInnerFilters(fs...)
	nf.SetOperation(op)
	nf.SetName(f.Name)
	nf.SetValue(f.Value)
	nf.SetKey(f.Key)
	return nf, nil
}

func (f *filter) fromNetmap(nf *netmap.Filter) {
	f.Name = nf.Name()
	f.Key = nf.Key()
	f.Value = nf.Value()
	switch nf.Operation() {
	case netmap.OpEQ:
		f.Op = "EQ"
	case netmap.OpNE:
		f.Op = "NE"
	case netmap.OpGT:
		f.Op = "GT"
	case netmap.OpGE:
		f.Op = "GE"
	case netmap.OpLT:
		f.Op = "LT"
	case netmap.OpLE:
		f.Op = "LE"
	case netmap.OpAND:
		f.Op = "AND"
	case netmap.OpOR:
		f.Op = "OR"
	default:
		// do nothing
	}
	if nf.InnerFilters() != nil {
		f.Filters = make([]filter, len(nf.InnerFilters()))
		for i, sf := range nf.InnerFilters() {
			f.Filters[i].fromNetmap(sf)
		}
	}
}

func (s *selector) toNetmap() (*netmap.Selector, error) {
	var c netmap.Clause
	switch strings.ToUpper(s.Clause) {
	case "SAME":
		c = netmap.ClauseSame
	case "DISTINCT":
		c = netmap.ClauseDistinct
	case "":
	default:
		return nil, fmt.Errorf("%w: '%s'", ErrUnknownClause, s.Clause)
	}
	ns := new(netmap.Selector)
	ns.SetName(s.Name)
	ns.SetAttribute(s.Attribute)
	ns.SetCount(s.Count)
	ns.SetClause(c)
	ns.SetFilter(s.Filter)
	return ns, nil
}

func (s *selector) fromNetmap(ns *netmap.Selector) {
	s.Name = ns.Name()
	s.Filter = ns.Filter()
	s.Count = ns.Count()
	s.Attribute = ns.Attribute()
	switch ns.Clause() {
	case netmap.ClauseSame:
		s.Clause = "same"
	case netmap.ClauseDistinct:
		s.Clause = "distinct"
	default:
		// do nothing
	}
	s.Name = ns.Name()
}
