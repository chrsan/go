package ormap

import "github.com/chrsan/go/crdt"

type State struct {
	m map[interface{}][]*Element
}

func (s *State) Len() int {
	return len(s.m)
}

func (s *State) Get(key interface{}) *Element {
	if s.m == nil {
		return nil
	}
	elements, ok := s.m[key]
	if !ok {
		return nil
	}
	return elements[0]
}

func (s *State) GetElement(key interface{}, dot *crdt.Dot) *Element {
	if s.m == nil {
		return nil
	}
	elements, ok := s.m[key]
	if !ok {
		return nil
	}
	i, found := crdt.BinarySearch(len(elements), func(i int) int { return elements[i].Dot.Cmp(dot) })
	if !found {
		return nil
	}
	return elements[i]
}

func (s *State) Insert(key, value interface{}, dot crdt.Dot) Op {
	ins := &Element{value, dot}
	var elements []*Element
	if !initMap(s) {
		elements = s.m[key]
	}
	s.m[key] = []*Element{ins}
	var removedDots []crdt.Dot
	if len(elements) != 0 {
		removedDots = make([]crdt.Dot, len(elements))
		for i, e := range elements {
			removedDots[i] = e.Dot
			e = nil
		}
		elements = nil
	}
	return Op{key, ins, removedDots}
}

func (s *State) Remove(key interface{}) (Op, bool) {
	if s.m == nil {
		return Op{}, false
	}
	elements, ok := s.m[key]
	if !ok {
		return Op{}, false
	}
	delete(s.m, key)
	removedDots := make([]crdt.Dot, len(elements))
	for i, e := range elements {
		removedDots[i] = e.Dot
		e = nil
	}
	elements = nil
	return Op{key, nil, removedDots}, true
}

func (s *State) ExecuteOp(op Op) LocalOp {
	var elements []*Element
	if s.m != nil {
		elements = s.m[op.Key]
	}
	if len(elements) != 0 {
		delete(s.m, op.Key)
		elems := elements[:0]
		for _, e := range elements {
			if !crdt.DotExists(op.RemovedDots, e.Dot) {
				elems = append(elems, e)
			} else {
				e = nil
			}
		}
		elements = elems
	}
	if op.InsertedElement != nil {
		e := *op.InsertedElement
		if i, found := crdt.BinarySearch(len(elements), func(i int) int { return elements[i].Dot.Cmp(&e.Dot) }); !found {
			elements = append(elements, nil)
			copy(elements[i+1:], elements[i:])
			elements[i] = &e
		}
	}
	if len(elements) == 0 {
		return LocalOp{false, op.Key, nil}
	}
	initMap(s)
	s.m[op.Key] = elements
	return LocalOp{true, op.Key, elements[0].Value}
}

func (s *State) Clone() *State {
	if s.m == nil {
		return &State{}
	}
	m := make(map[interface{}][]*Element, len(s.m))
	for k, elements := range s.m {
		elems := make([]*Element, len(elements))
		copy(elems, elements)
		m[k] = elems
	}
	return &State{m}
}

func (s *State) Transform(f func(*Element) *Element) *State {
	if s.m == nil {
		return &State{}
	}
	m := make(map[interface{}][]*Element, len(s.m))
	for k, elements := range s.m {
		elems := make([]*Element, len(elements))
		for i, e := range elements {
			elems[i] = f(e)
		}
		m[k] = elems
	}
	return &State{m}
}

func (s *State) Entries(f func(e Entry)) {
	if s.m == nil {
		return
	}
	for k, elements := range s.m {
		f(Entry{k, elements[0].Value})
	}
}

func (s *State) Eq(state *State) bool {
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

func initMap(s *State) bool {
	if s.m != nil {
		return false
	}
	s.m = map[interface{}][]*Element{}
	return true
}
