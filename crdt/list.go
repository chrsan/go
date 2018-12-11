package crdt

type List struct {
	siteID  SiteID
	state   *listState
	summary *Summary
}

func NewList(siteID SiteID) *List {
	checkSiteID(siteID)
	return &List{siteID, &listState{}, NewSummary()}
}

func (l *List) SiteID() SiteID {
	return l.siteID
}

func (l *List) Len() int {
	return len(l.state.elements)
}

func (l *List) Get(i int) interface{} {
	return l.state.elements[i].Value
}

func (l *List) Push(value interface{}) ListInsertOp {
	dot := l.summary.Dot(l.siteID)
	return l.state.push(value, &dot)
}

func (l *List) Insert(i int, value interface{}) ListInsertOp {
	dot := l.summary.Dot(l.siteID)
	return l.state.insert(i, value, &dot)
}

func (l *List) Pop() (interface{}, ListRemoveOp) {
	return l.state.pop()
}

func (l *List) Remove(i int) (interface{}, ListRemoveOp) {
	return l.state.remove(i)
}

func (l *List) ExecuteOp(op ListOp) (LocalListOp, bool) {
	for _, dot := range op.InsertedDots() {
		l.summary.Insert(&dot)
	}
	return l.state.executeOp(op)
}

func (l *List) Values(f func(interface{})) {
	l.state.values(f)
}

func (l *List) Replicate(siteID SiteID) *List {
	checkSiteID(siteID)
	return &List{siteID, l.state.clone(), l.summary.Clone()}
}

func (l *List) Eq(list *List) bool {
	return l.state.eq(list.state) && l.summary.Eq(list.summary)
}

type ListElement struct {
	UID   UID
	Value interface{}
}

type ListOp interface {
	InsertedDots() []Dot
	InsertedElement() (ListElement, bool)
	RemovedUID() (UID, bool)
	Validate(siteID SiteID) bool
	isOp()
}

type ListInsertOp struct {
	Element ListElement
}

func (i ListInsertOp) InsertedDots() []Dot {
	return []Dot{i.Element.UID.Dot}
}

func (i ListInsertOp) InsertedElement() (ListElement, bool) {
	return i.Element, true
}

func (i ListInsertOp) RemovedUID() (UID, bool) {
	return UID{}, false
}

func (i ListInsertOp) Validate(siteID SiteID) bool {
	if i.Element.UID.Dot.SiteID != siteID {
		return false
	}
	return true
}

func (ListInsertOp) isOp() {}

type ListRemoveOp = UID

func (r ListRemoveOp) InsertedDots() []Dot {
	return nil
}

func (r ListRemoveOp) InsertedElement() (ListElement, bool) {
	return ListElement{}, false
}

func (r ListRemoveOp) RemovedUID() (UID, bool) {
	return r, true
}

func (r ListRemoveOp) Validate(siteID SiteID) bool {
	return true
}

func (ListRemoveOp) isOp() {}

type LocalListOp interface {
	isLocalOp()
}

type LocalListInsertOp struct {
	Index int
	Value interface{}
}

func (LocalListInsertOp) isLocalOp() {}

type LocalListRemoveOp int

func (LocalListRemoveOp) isLocalOp() {}

type listState struct {
	elements []*ListElement
}

func (s *listState) push(value interface{}, dot *Dot) ListInsertOp {
	uid1 := &MinUID
	if len(s.elements) != 0 {
		uid1 = &s.elements[len(s.elements)-1].UID
	}
	uid := UIDBetween(uid1, &MaxUID, dot)
	e := ListElement{uid, value}
	s.elements = append(s.elements, &e)
	return ListInsertOp{e}
}

func (s *listState) insert(i int, value interface{}, dot *Dot) ListInsertOp {
	uid1, uid2 := &MinUID, &MaxUID
	if i != 0 {
		uid1 = &s.elements[i-1].UID
	}
	if i != len(s.elements) {
		uid2 = &s.elements[i].UID
	}
	uid := UIDBetween(uid1, uid2, dot)
	e := ListElement{uid, value}
	s.elements = append(s.elements, nil)
	copy(s.elements[i+1:], s.elements[i:])
	s.elements[i] = &e
	return ListInsertOp{e}
}

func (s *listState) pop() (interface{}, ListRemoveOp) {
	e := s.elements[0]
	v, u := e.Value, e.UID
	s.elements[0] = nil
	s.elements = s.elements[1:]
	return v, u
}

func (s *listState) remove(i int) (interface{}, ListRemoveOp) {
	e := s.elements[i]
	v, u := e.Value, e.UID
	copy(s.elements[i:], s.elements[i+1:])
	s.elements[len(s.elements)-1] = nil
	s.elements = s.elements[:len(s.elements)-1]
	return v, u
}

func (s *listState) index(uid *UID) (int, bool) {
	return BinarySearch(len(s.elements), func(i int) int {
		e := s.elements[i]
		return (&e.UID).Cmp(uid)
	})
}

func (s *listState) executeOp(op ListOp) (LocalListOp, bool) {
	switch op := op.(type) {
	case ListInsertOp:
		e := op.Element
		i, found := s.index(&e.UID)
		if found {
			return LocalListInsertOp{}, false
		}
		if i == len(s.elements) {
			s.elements = append(s.elements, &e)
		} else {
			s.elements = append(s.elements, nil)
			copy(s.elements[i+1:], s.elements[i:])
			s.elements[i] = &e
		}
		return LocalListInsertOp{Index: i, Value: e.Value}, true
	case ListRemoveOp:
		i, found := s.index(&op)
		if !found {
			return LocalListRemoveOp(-1), false
		}
		copy(s.elements[i:], s.elements[i+1:])
		s.elements[len(s.elements)-1] = nil
		s.elements = s.elements[:len(s.elements)-1]
		return LocalListRemoveOp(i), true
	}
	panic(op)
}

func (s *listState) clone() *listState {
	if len(s.elements) == 0 {
		return &listState{}
	}
	elements := make([]*ListElement, len(s.elements))
	copy(elements, s.elements)
	return &listState{elements}
}

func (s *listState) values(f func(interface{})) {
	for _, e := range s.elements {
		f(e.Value)
	}
}

func (s *listState) eq(state *listState) bool {
	if len(s.elements) != len(state.elements) {
		return false
	}
	for i, e := range s.elements {
		if (&e.UID).Cmp(&state.elements[i].UID) != 0 {
			return false
		}
	}
	return true
}
