package crdt

type Map struct {
	siteID  SiteID
	state   *mapState
	summary *Summary
}

func NewMap(siteID SiteID) *Map {
	checkSiteID(siteID)
	return &Map{siteID, &mapState{}, NewSummary()}
}

func (m *Map) SiteID() SiteID {
	return m.siteID
}

func (m *Map) Len() int {
	return len(m.state.m)
}

func (m *Map) Contains(key interface{}) bool {
	if m.state.m == nil {
		return false
	}
	if _, ok := m.state.m[key]; !ok {
		return false
	}
	return true
}

func (m *Map) Get(key interface{}) interface{} {
	e := m.state.get(key)
	if e == nil {
		return nil
	}
	return e.Value
}

func (m *Map) Insert(key, value interface{}) MapOp {
	dot := m.summary.Dot(m.siteID)
	return m.state.insert(key, value, dot)
}

func (m *Map) Remove(key interface{}) (MapOp, bool) {
	return m.state.remove(key)
}

func (m *Map) ExecuteOp(op MapOp) (LocalMapOp, bool) {
	if op.InsertedElement != nil {
		m.summary.Insert(&op.InsertedElement.Dot)
	}
	return m.state.executeOp(op)
}

func (m *Map) Values() []MapEntry {
	return m.state.values()
}

func (m *Map) Replicate(siteID SiteID) *Map {
	checkSiteID(siteID)
	return &Map{siteID, m.state.clone(), m.summary.Clone()}
}

func (m *Map) Eq(n *Map) bool {
	return m.state.eq(n.state) && m.summary.Eq(n.summary)
}

type MapEntry struct {
	Key, Value interface{}
}

type MapElement struct {
	Value interface{}
	Dot   Dot
}

type MapOp struct {
	Key             interface{}
	InsertedElement *MapElement
	RemovedDots     []Dot
}

type LocalMapOp struct {
	IsInsert   bool
	Key, Value interface{}
}

type mapState struct {
	m map[interface{}][]*MapElement
}

func (s *mapState) get(key interface{}) *MapElement {
	if s.m == nil {
		return nil
	}
	elements, ok := s.m[key]
	if !ok {
		return nil
	}
	return elements[0]
}

func (s *mapState) getElement(key interface{}, dot Dot) *MapElement {
	if s.m == nil {
		return nil
	}
	elements, ok := s.m[key]
	if !ok {
		return nil
	}
	i, found := BinarySearch(len(elements), func(i int) int { return elements[i].Dot.Cmp(dot) })
	if !found {
		return nil
	}
	return elements[i]
}

func (s *mapState) insert(key, value interface{}, dot Dot) MapOp {
	ins := &MapElement{value, dot}
	var elements []*MapElement
	if !initMap(s) {
		elements = s.m[key]
	}
	s.m[key] = []*MapElement{ins}
	var removedDots []Dot
	if len(elements) != 0 {
		removedDots = make([]Dot, len(elements))
		for i, e := range elements {
			removedDots[i] = e.Dot
		}
	}
	return MapOp{key, ins, removedDots}
}

func (s *mapState) remove(key interface{}) (MapOp, bool) {
	if s.m == nil {
		return MapOp{}, false
	}
	elements, ok := s.m[key]
	if !ok {
		return MapOp{}, false
	}
	delete(s.m, key)
	removedDots := make([]Dot, len(elements))
	for i, e := range elements {
		removedDots[i] = e.Dot
	}
	return MapOp{key, nil, removedDots}, true
}

func (s *mapState) executeOp(op MapOp) (LocalMapOp, bool) {
	var elements []*MapElement
	if s.m != nil {
		elements = s.m[op.Key]
	}
	if len(elements) != 0 {
		delete(s.m, op.Key)
		elems := elements[:0]
		for _, e := range elements {
			if !dotExists(op.RemovedDots, e.Dot) {
				elems = append(elems, e)
			}
		}
		elements = elems
	}
	if op.InsertedElement != nil {
		e := *op.InsertedElement
		if i, found := BinarySearch(len(elements), func(i int) int { return elements[i].Dot.Cmp(e.Dot) }); !found {
			elements = append(elements, nil)
			copy(elements[i+1:], elements[i:])
			elements[i] = &e
		}
	}
	if len(elements) == 0 {
		return LocalMapOp{false, op.Key, nil}, true
	}
	initMap(s)
	s.m[op.Key] = elements
	return LocalMapOp{true, op.Key, elements[0].Value}, true
}

func (s *mapState) clone() *mapState {
	if s.m == nil {
		return &mapState{}
	}
	m := make(map[interface{}][]*MapElement, len(s.m))
	for k, elements := range s.m {
		elems := make([]*MapElement, len(elements))
		copy(elems, elements)
		m[k] = elems
	}
	return &mapState{m}
}

func (s *mapState) values() []MapEntry {
	if s.m == nil {
		return nil
	}
	es, i := make([]MapEntry, len(s.m)), 0
	for k, elements := range s.m {
		es[i] = MapEntry{k, elements[0].Value}
	}
	return es
}

func (s *mapState) eq(state *mapState) bool {
	if len(s.m) == 0 && len(state.m) == 0 {
		return true
	}
	if len(s.m) != len(state.m) {
		return false
	}
	for k, elements1 := range s.m {
		elements2, ok := state.m[k]
		if !ok {
			return false
		}
		if len(elements1) != len(elements2) {
			return false
		}
		for i, e := range elements1 {
			if e.Dot != elements2[i].Dot {
				return false
			}
		}
	}
	return true
}

func initMap(s *mapState) bool {
	if s.m != nil {
		return false
	}
	s.m = map[interface{}][]*MapElement{}
	return true
}
