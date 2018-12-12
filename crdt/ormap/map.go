package ormap

import "github.com/chrsan/go/crdt"

type Map struct {
	siteID  crdt.SiteID
	state   *State
	summary *crdt.Summary
}

func New(siteID crdt.SiteID) *Map {
	crdt.CheckSiteID(siteID)
	return &Map{siteID, &State{}, crdt.NewSummary()}
}

func (m *Map) SiteID() crdt.SiteID {
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
	e := m.state.Get(key)
	if e == nil {
		return nil
	}
	return e.Value
}

func (m *Map) Insert(key, value interface{}) Op {
	dot := m.summary.Dot(m.siteID)
	return m.state.Insert(key, value, dot)
}

func (m *Map) Remove(key interface{}) (Op, bool) {
	return m.state.Remove(key)
}

func (m *Map) ExecuteOp(op Op) LocalOp {
	if op.InsertedElement != nil {
		m.summary.Insert(&op.InsertedElement.Dot)
	}
	return m.state.ExecuteOp(op)
}

func (m *Map) Entries(f func(Entry)) {
	m.state.Entries(f)
}

func (m *Map) Replicate(siteID crdt.SiteID) *Map {
	crdt.CheckSiteID(siteID)
	return &Map{siteID, m.state.Clone(), m.summary.Clone()}
}

func (m *Map) Eq(n *Map) bool {
	return m.state.Eq(n.state) && m.summary.Eq(n.summary)
}

type Element struct {
	Value interface{}
	Dot   crdt.Dot
}

type Entry struct {
	Key, Value interface{}
}
