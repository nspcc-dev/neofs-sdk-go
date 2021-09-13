package eacl

import (
	"bytes"
	"errors"

	"go.uber.org/zap"
)

// Validator is a tool that calculates
// the action on a request according
// to the extended ACL rule table.
type Validator struct {
	*cfg
}

// NewValidator creates and initializes a new Validator using options.
func NewValidator(opts ...Option) *Validator {
	cfg := defaultCfg()

	for i := range opts {
		opts[i](cfg)
	}

	return &Validator{
		cfg: cfg,
	}
}

// CalculateAction calculates action on the request according
// to its information represented in ValidationUnit.
//
// The action is calculated according to the application of
// eACL table of rules to the request.
//
// If the eACL table is not available at the time of the call,
// ActionUnknown is returned.
//
// If no matching table entry is found, ActionAllow is returned.
func (v *Validator) CalculateAction(unit *ValidationUnit) Action {
	var (
		err   error
		table *Table
	)

	if unit.bearer != nil {
		table = NewTableFromV2(unit.bearer.GetBody().GetEACL())
	} else {
		// get eACL table by container ID
		table, err = v.storage.GetEACL(unit.cid)
		if err != nil {
			if errors.Is(err, ErrEACLNotFound) {
				return ActionAllow
			}

			v.logger.Error("could not get eACL table",
				zap.String("error", err.Error()),
			)

			return ActionUnknown
		}
	}

	return tableAction(unit, table)
}

// tableAction calculates action on the request based on the eACL rules.
func tableAction(unit *ValidationUnit, table *Table) Action {
	for _, record := range table.Records() {
		// check type of operation
		if record.Operation() != unit.op {
			continue
		}

		// check target
		if !targetMatches(unit, record) {
			continue
		}

		// check headers
		switch val := matchFilters(unit.hdrSrc, record.Filters()); {
		case val < 0:
			// headers of some type could not be composed => allow
			return ActionAllow
		case val == 0:
			return record.Action()
		}
	}

	return ActionAllow
}

// returns:
//  - positive value if no matching header is found for at least one filter;
//  - zero if at least one suitable header is found for all filters;
//  - negative value if the headers of at least one filter cannot be obtained.
func matchFilters(hdrSrc TypedHeaderSource, filters []*Filter) int {
	matched := 0

	for _, filter := range filters {
		headers, ok := hdrSrc.HeadersOfType(filter.From())
		if !ok {
			return -1
		}

		// get headers of filtering type
		for _, header := range headers {
			// prevent NPE
			if header == nil {
				continue
			}

			// check header name
			if header.Key() != filter.Key() {
				continue
			}

			// get match function
			matchFn, ok := mMatchFns[filter.Matcher()]
			if !ok {
				continue
			}

			// check match
			if !matchFn(header, filter) {
				continue
			}

			// increment match counter
			matched++

			break
		}
	}

	return len(filters) - matched
}

// returns true if one of ExtendedACLTarget has
// suitable target OR suitable public key.
func targetMatches(unit *ValidationUnit, record *Record) bool {
	for _, target := range record.Targets() {
		// check public key match
		if pubs := target.BinaryKeys(); len(pubs) != 0 {
			for _, key := range pubs {
				if bytes.Equal(key, unit.key) {
					return true
				}
			}
			continue
		}

		// check target group match
		if unit.role == target.Role() {
			return true
		}
	}

	return false
}

// Maps match type to corresponding function.
var mMatchFns = map[Match]func(Header, *Filter) bool{
	MatchStringEqual: func(header Header, filter *Filter) bool {
		return header.Value() == filter.Value()
	},

	MatchStringNotEqual: func(header Header, filter *Filter) bool {
		return header.Value() != filter.Value()
	},
}
