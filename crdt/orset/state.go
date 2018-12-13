package orset

import "github.com/chrsan/go/crdt"

type State struct {
	s map[interface{}][]crdt.Dot
}

func (s *State) Insert(value interface{}, dot *crdt.Dot) Op {
	var removedDots []crdt.Dot
	if !initSet(s) {
		removedDots = s.s[value]
	}
	s.s[value] = []crdt.Dot{*dot}
	return Op{OpInsert, value, dot, removedDots}
}

func (s *State) Remove(value interface{}) (Op, bool) {
	if s.s == nil {
		return Op{}, false
	}
	removedDots, ok := s.s[value]
	if !ok {
		return Op{}, false
	}
	delete(s.s, value)
	return Op{OpRemove, value, nil, removedDots}, true
}

func (s *State) ExecuteOp(op Op) (LocalOp, bool) {
	var dots []crdt.Dot
	if s.s != nil {
		dots = s.s[op.Value]
	}
	existsBefore := len(dots) != 0
	if existsBefore {
		delete(s.s, op.Value)
		ds := dots[:0]
		for _, d := range dots {
			if !crdt.DotExists(op.RemovedDots, d) {
				ds = append(ds, d)
			}
		}
		dots = ds
	}
	if op.InsertedDot != nil {
		if i, found := crdt.BinarySearch(len(dots), func(i int) int { return dots[i].Cmp(op.InsertedDot) }); !found {
			dots = append(dots, crdt.Dot{})
			copy(dots[i+1:], dots[i:])
			dots[i] = *op.InsertedDot
		}
	}
	if existsBefore && len(dots) != 0 {
		initSet(s)
		s.s[op.Value] = dots
	} else if len(dots) != 0 {
		initSet(s)
		s.s[op.Value] = dots
		return LocalOp{OpInsert, op.Value}, true
	} else if existsBefore {
		return LocalOp{OpRemove, op.Value}, true
	}
	return LocalOp{}, false
}

func (s *State) Clone() *State {
	if s.s == nil {
		return &State{}
	}
	set := make(map[interface{}][]crdt.Dot, len(s.s))
	for k, dots := range s.s {
		ds := make([]crdt.Dot, len(dots))
		copy(ds, dots)
		set[k] = ds
	}
	return &State{set}
}

func (s *State) Values(f func(interface{})) {
	if s.s == nil {
		return
	}
	for k := range s.s {
		f(k)
	}
}

func (s *State) Eq(state *State) bool {
	if len(s.s) == 0 && len(state.s) == 0 {
		return true
	}
	if len(s.s) != len(state.s) {
		return false
	}
	for k, dots1 := range s.s {
		dots2, ok := state.s[k]
		if !ok {
			return false
		}
		if len(dots1) != len(dots2) {
			return false
		}
		for i, d := range dots1 {
			if dots2[i] != d {
				return false
			}
		}
	}
	return true
}

func initSet(s *State) bool {
	if s.s != nil {
		return false
	}
	s.s = map[interface{}][]crdt.Dot{}
	return true
}
