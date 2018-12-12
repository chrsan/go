package orset

import "github.com/chrsan/go/crdt"

type Set struct {
	siteID  crdt.SiteID
	state   *State
	summary *crdt.Summary
}

func New(siteID crdt.SiteID) *Set {
	crdt.CheckSiteID(siteID)
	return &Set{siteID, &State{}, crdt.NewSummary()}
}

func (s *Set) SiteID() crdt.SiteID {
	return s.siteID
}

func (s *Set) Len() int {
	return len(s.state.s)
}

func (s *Set) Contains(value interface{}) bool {
	if s.state.s == nil {
		return false
	}
	if _, ok := s.state.s[value]; !ok {
		return false
	}
	return true
}

func (s *Set) Insert(value interface{}) Op {
	dot := s.summary.Dot(s.siteID)
	return s.state.Insert(value, &dot)
}

func (s *Set) Remove(value interface{}) (Op, bool) {
	return s.state.Remove(value)
}

func (s *Set) ExecuteOp(op Op) (LocalOp, bool) {
	if op.InsertedDot != nil {
		s.summary.Insert(op.InsertedDot)
	}
	return s.state.ExecuteOp(op)
}

func (s *Set) Values(f func(interface{})) {
	s.state.Values(f)
}

func (s *Set) Replicate(siteID crdt.SiteID) *Set {
	crdt.CheckSiteID(siteID)
	return &Set{siteID, s.state.Clone(), s.summary.Clone()}
}

func (s *Set) Eq(set *Set) bool {
	return s.state.Eq(set.state) && s.summary.Eq(set.summary)
}
