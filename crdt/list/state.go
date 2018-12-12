package list

import "github.com/chrsan/go/crdt"

type State struct {
	elements []*Element
}

func (s *State) Len() int {
	return len(s.elements)
}

func (s *State) Get(i int) *Element {
	return s.elements[i]
}

func (s *State) Push(value interface{}, dot *crdt.Dot) InsertOp {
	uid1 := &crdt.MinUID
	if len(s.elements) != 0 {
		uid1 = &s.elements[len(s.elements)-1].UID
	}
	uid := crdt.UIDBetween(uid1, &crdt.MaxUID, dot)
	e := Element{uid, value}
	s.elements = append(s.elements, &e)
	return InsertOp{e}
}

func (s *State) Insert(i int, value interface{}, dot *crdt.Dot) InsertOp {
	uid1, uid2 := &crdt.MinUID, &crdt.MaxUID
	if i != 0 {
		uid1 = &s.elements[i-1].UID
	}
	if i != len(s.elements) {
		uid2 = &s.elements[i].UID
	}
	uid := crdt.UIDBetween(uid1, uid2, dot)
	e := Element{uid, value}
	s.elements = append(s.elements, nil)
	copy(s.elements[i+1:], s.elements[i:])
	s.elements[i] = &e
	return InsertOp{e}
}

func (s *State) Pop() (interface{}, RemoveOp) {
	e := s.elements[0]
	v, u := e.Value, e.UID
	s.elements[0] = nil
	s.elements = s.elements[1:]
	return v, RemoveOp{u}
}

func (s *State) Remove(i int) (interface{}, RemoveOp) {
	e := s.elements[i]
	v, u := e.Value, e.UID
	copy(s.elements[i:], s.elements[i+1:])
	s.elements[len(s.elements)-1] = nil
	s.elements = s.elements[:len(s.elements)-1]
	return v, RemoveOp{u}
}

func (s *State) Index(uid *crdt.UID) (int, bool) {
	return crdt.BinarySearch(len(s.elements), func(i int) int {
		e := s.elements[i]
		return (&e.UID).Cmp(uid)
	})
}

func (s *State) ExecuteOp(op Op) (LocalOp, bool) {
	switch op := op.(type) {
	case InsertOp:
		e := op.Element
		i, found := s.Index(&e.UID)
		if found {
			return LocalInsertOp{}, false
		}
		if i == len(s.elements) {
			s.elements = append(s.elements, &e)
		} else {
			s.elements = append(s.elements, nil)
			copy(s.elements[i+1:], s.elements[i:])
			s.elements[i] = &e
		}
		return LocalInsertOp{Index: i, Value: e.Value}, true
	case RemoveOp:
		i, found := s.Index(&op.UID)
		if !found {
			return LocalRemoveOp(-1), false
		}
		copy(s.elements[i:], s.elements[i+1:])
		s.elements[len(s.elements)-1] = nil
		s.elements = s.elements[:len(s.elements)-1]
		return LocalRemoveOp(i), true
	}
	panic(op)
}

func (s *State) Clone() *State {
	if len(s.elements) == 0 {
		return &State{}
	}
	elements := make([]*Element, len(s.elements))
	copy(elements, s.elements)
	return &State{elements}
}

func (s *State) Transform(f func(*Element) *Element) *State {
	if len(s.elements) == 0 {
		return &State{}
	}
	elements := make([]*Element, len(s.elements))
	for i, e := range s.elements {
		elements[i] = f(e)
	}
	return &State{elements}
}

func (s *State) Values(f func(interface{})) {
	for _, e := range s.elements {
		f(e.Value)
	}
}

func (s *State) Eq(state *State) bool {
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
