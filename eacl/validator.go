package eacl

import (
	"bytes"
)

// Validator is a tool that calculates
// the action on a request according
// to the extended ACL rule table.
type Validator struct {
}

// NewValidator creates and initializes a new Validator using options.
func NewValidator() *Validator {
	return &Validator{}
}

// CalculateAction calculates action on the request according
// to its information represented in ValidationUnit.
//
// The action is calculated according to the application of
// eACL table of rules to the request.
//
// If no matching table entry is found, ActionAllow is returned.
func (v *Validator) CalculateAction(unit *ValidationUnit) Action {
	for _, record := range unit.table.Records() {
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
