package list

import "github.com/chrsan/go/crdt"

type List struct {
	siteID  crdt.SiteID
	state   *State
	summary *crdt.Summary
}

func New(siteID crdt.SiteID) *List {
	crdt.CheckSiteID(siteID)
	return &List{siteID, &State{}, crdt.NewSummary()}
}

func (l *List) SiteID() crdt.SiteID {
	return l.siteID
}

func (l *List) Len() int {
	return l.state.Len()
}

func (l *List) Get(i int) interface{} {
	return l.state.elements[i].Value
}

func (l *List) Push(value interface{}) InsertOp {
	dot := l.summary.Dot(l.siteID)
	return l.state.Push(value, &dot)
}

func (l *List) Insert(i int, value interface{}) InsertOp {
	dot := l.summary.Dot(l.siteID)
	return l.state.Insert(i, value, &dot)
}

func (l *List) Pop() (interface{}, RemoveOp) {
	return l.state.Pop()
}

func (l *List) Remove(i int) (interface{}, RemoveOp) {
	return l.state.Remove(i)
}

func (l *List) ExecuteOp(op Op) (LocalOp, bool) {
	e, ok := op.InsertedElement()
	if ok {
		l.summary.Insert(&e.UID.Dot)
	}
	return l.state.ExecuteOp(op)
}

func (l *List) Values(f func(interface{})) {
	l.state.Values(f)
}

func (l *List) Replicate(siteID crdt.SiteID) *List {
	crdt.CheckSiteID(siteID)
	return &List{siteID, l.state.Clone(), l.summary.Clone()}
}

func (l *List) Eq(list *List) bool {
	return l.state.Eq(list.state) && l.summary.Eq(list.summary)
}

type Element struct {
	UID   crdt.UID
	Value interface{}
}
