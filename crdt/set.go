package crdt

import (
	"fmt"
)

type Set struct {
	siteID  SiteID
	state   *setState
	summary *Summary
}

func NewSet(siteID SiteID) *Set {
	checkSiteID(siteID)
	return &Set{siteID, &setState{}, NewSummary()}
}

func (s *Set) SiteID() SiteID {
	return s.siteID
}

func (s *Set) Len() int {
	return len(s.state.set)
}

func (s *Set) Contains(value interface{}) bool {
	if s.state.set == nil {
		return false
	}
	if _, ok := s.state.set[value]; !ok {
		return false
	}
	return true
}

func (s *Set) Insert(value interface{}) SetOp {
	dot := s.summary.Dot(s.siteID)
	return s.state.insert(value, &dot)
}

func (s *Set) Remove(value interface{}) (SetOp, bool) {
	return s.state.remove(value)
}

func (s *Set) ExecuteOp(op SetOp) (LocalSetOp, bool) {
	if op.InsertedDot != nil {
		s.summary.Insert(op.InsertedDot)
	}
	return s.state.executeOp(op)
}

func (s *Set) Values(f func(interface{})) {
	s.state.values(f)
}

func (s *Set) Replicate(siteID SiteID) *Set {
	checkSiteID(siteID)
	return &Set{siteID, s.state.clone(), s.summary.Clone()}
}

func (s *Set) Eq(set *Set) bool {
	return s.state.eq(set.state) && s.summary.Eq(set.summary)
}

type SetOpKind int

const (
	SetOpInsert SetOpKind = iota
	SetOpRemove
)

type SetOp struct {
	Kind        SetOpKind
	Value       interface{}
	InsertedDot *Dot
	RemovedDots []Dot
}

func (s SetOp) Validate(siteID SiteID) error {
	if s.InsertedDot != nil && s.InsertedDot.SiteID != siteID {
		return fmt.Errorf("Invalid op: %d != %d", s.InsertedDot.SiteID, siteID)
	}
	return nil
}

type LocalSetOp struct {
	Kind  SetOpKind
	Value interface{}
}

type setState struct {
	set map[interface{}][]Dot
}

func (s *setState) insert(value interface{}, dot *Dot) SetOp {
	var removedDots []Dot
	if !initSet(s) {
		removedDots = s.set[value]
	}
	s.set[value] = []Dot{*dot}
	return SetOp{SetOpInsert, value, dot, removedDots}
}

func (s *setState) remove(value interface{}) (SetOp, bool) {
	if s.set == nil {
		return SetOp{}, false
	}
	removedDots, ok := s.set[value]
	if !ok {
		return SetOp{}, false
	}
	delete(s.set, value)
	return SetOp{SetOpRemove, value, nil, removedDots}, true
}

func (s *setState) executeOp(op SetOp) (LocalSetOp, bool) {
	var dots []Dot
	if s.set != nil {
		dots = s.set[op.Value]
	}
	existsBefore := len(dots) != 0
	if existsBefore {
		delete(s.set, op.Value)
		ds := dots[:0]
		for _, d := range dots {
			if !dotExists(op.RemovedDots, d) {
				ds = append(ds, d)
			}
		}
		dots = ds
	}
	if op.InsertedDot != nil {
		if i, found := BinarySearch(len(dots), func(i int) int { return dots[i].Cmp(*op.InsertedDot) }); !found {
			dots = append(dots, Dot{})
			copy(dots[i+1:], dots[i:])
			dots[i] = *op.InsertedDot
		}
	}
	if existsBefore && len(dots) != 0 {
		initSet(s)
		s.set[op.Value] = dots
	} else if len(dots) != 0 {
		initSet(s)
		s.set[op.Value] = dots
		return LocalSetOp{SetOpInsert, op.Value}, true
	} else if existsBefore {
		return LocalSetOp{SetOpRemove, op.Value}, true
	}
	return LocalSetOp{}, false
}

func (s *setState) clone() *setState {
	if s.set == nil {
		return &setState{}
	}
	set := make(map[interface{}][]Dot, len(s.set))
	for k, dots := range s.set {
		ds := make([]Dot, len(dots))
		copy(ds, dots)
		set[k] = ds
	}
	return &setState{set}
}

func (s *setState) values(f func(interface{})) {
	if s.set == nil {
		return
	}
	for k := range s.set {
		f(k)
	}
}

func (s *setState) eq(state *setState) bool {
	if len(s.set) == 0 && len(state.set) == 0 {
		return true
	}
	if len(s.set) != len(state.set) {
		return false
	}
	for k, dots1 := range s.set {
		dots2, ok := state.set[k]
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

func initSet(s *setState) bool {
	if s.set != nil {
		return false
	}
	s.set = map[interface{}][]Dot{}
	return true
}
