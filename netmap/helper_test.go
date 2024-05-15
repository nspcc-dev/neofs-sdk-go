package netmap

func newFilter(name string, k, v string, op FilterOp, fs ...Filter) (f Filter) {
	return Filter{
		name: name,
		key:  k,
		op:   op,
		val:  v,
		subs: fs,
	}
}

func newSelector(name string, attr string, count uint32, filter string, clause func(*Selector)) (s Selector) {
	s.SetName(name)
	s.SelectByBucketAttribute(attr)
	s.SetNumberOfNodes(count)
	clause(&s)
	s.SetFilterName(filter)
	return s
}

func newPlacementPolicy(bf uint32, rs []ReplicaDescriptor, ss []Selector, fs []Filter) (p PlacementPolicy) {
	p.SetContainerBackupFactor(bf)
	p.SetReplicas(rs)
	p.SetSelectors(ss)
	p.SetFilters(fs)
	return p
}

func newReplica(c uint32, s string) (r ReplicaDescriptor) {
	r.SetNumberOfObjects(c)
	r.SetSelectorName(s)
	return r
}

func nodeInfoFromAttributes(props ...string) (n NodeInfo) {
	for i := 0; i < len(props); i += 2 {
		n.SetAttribute(props[i], props[i+1])
	}

	return
}
