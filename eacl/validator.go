package eacl

import (
	"bytes"
	"math/big"
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
// Second return value is true iff the action was produced by a matching entry.
//
// If no matching table entry is found or some filters are missing,
// ActionAllow is returned and the second return value is false.
//
// Note that if some rule imposes requirements on the format of values (like
// numeric), but they do not comply with it - such a rule does not match.
func (v *Validator) CalculateAction(unit *ValidationUnit) (Action, bool, error) {
	for _, record := range unit.table.Records() {
		// check type of operation
		if record.Operation() != unit.op {
			continue
		}

		// check target
		if !targetMatches(unit, &record) {
			continue
		}

		// check headers
		switch val, err := matchFilters(unit.hdrSrc, record.Filters()); {
		case err != nil:
			return ActionDeny, false, err
		case val < 0:
			// headers of some type could not be composed => allow
			return ActionAllow, false, nil
		case val == 0:
			return record.Action(), true, nil
		}
	}

	return ActionAllow, false, nil
}

// returns:
//   - positive value if no matching header is found for at least one filter;
//   - zero if at least one suitable header is found for all filters;
//   - negative value if the headers of at least one filter cannot be obtained.
func matchFilters(hdrSrc TypedHeaderSource, filters []Filter) (int, error) {
	matched := 0
	var nv, nf big.Int

nextFilter:
	for _, filter := range filters {
		headers, ok, err := hdrSrc.HeadersOfType(filter.From())
		if err != nil {
			return 0, err
		}
		if !ok {
			return -1, nil
		}

		m := filter.Matcher()
		if m == MatchNumGT || m == MatchNumGE || m == MatchNumLT || m == MatchNumLE {
			if _, ok = nf.SetString(filter.Value(), 10); !ok {
				continue
			}
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

			// check match
			switch m {
			default:
				continue
			case MatchNotPresent:
				continue nextFilter
			case MatchStringEqual:
				if header.Value() != filter.Value() {
					continue
				}
			case MatchStringNotEqual:
				if header.Value() == filter.Value() {
					continue
				}
			case MatchNumGT, MatchNumGE, MatchNumLT, MatchNumLE:
				// TODO: big math simplifies coding but almost always not efficient
				//  enough, try to optimize
				if _, ok = nv.SetString(header.Value(), 10); !ok {
					continue
				}
				switch nv.Cmp(&nf) {
				default:
					continue // should never happen but just in case
				case -1:
					if m == MatchNumGT || m == MatchNumGE {
						continue
					}
				case 0:
					if m == MatchNumGT || m == MatchNumLT {
						continue
					}
				case 1:
					if m == MatchNumLT || m == MatchNumLE {
						continue
					}
				}
			}

			// increment match counter
			matched++

			break
		}

		if m == MatchNotPresent {
			matched++
		}
	}

	return len(filters) - matched, nil
}

// returns true if one of ExtendedACLTarget has
// suitable target OR suitable public key.
func targetMatches(unit *ValidationUnit, record *Record) bool {
	var checkRole bool

	for _, target := range record.Targets() {
		if target.Role() == RoleSystem {
			// system role access modifications have been deprecated
			continue
		}

		checkRole = true

		// check public key match
		if pubs := target.binaryKeys(); len(pubs) != 0 {
			checkRole = false

			for _, key := range pubs {
				if bytes.Equal(key, unit.key) {
					return true
				}
			}
		}

		// check accounts key match
		if accounts := target.Accounts(); len(accounts) != 0 {
			checkRole = false

			for _, account := range accounts {
				if bytes.Equal(account[:], unit.account) {
					return true
				}
			}
		}

		// check target group match
		if checkRole && unit.role == target.Role() {
			return true
		}
	}

	return false
}
