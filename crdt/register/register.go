package register

import (
	"github.com/chrsan/go/crdt"
	"github.com/google/btree"
)

type Register struct {
	siteID   crdt.SiteID
	elements *btree.BTree
	summary  *crdt.Summary
}

func New(siteID crdt.SiteID, value interface{}) *Register {
	crdt.CheckSiteID(siteID)
	elements := btree.New(5)
	summary := crdt.NewSummary()
	counter := crdt.Counter(1)
	summary.Insert(&crdt.Dot{SiteID: siteID, Counter: counter})
	elements.ReplaceOrInsert(newElem(siteID, value, counter))
	return &Register{siteID, elements, summary}
}

func (r *Register) SiteID() crdt.SiteID {
	return r.siteID
}

func (r *Register) Counter(siteID crdt.SiteID) crdt.Counter {
	return r.summary.Counter(siteID)
}

func (r *Register) Get() interface{} {
	e := r.elements.Min().(elem)
	return e.value
}

func (r *Register) Update(value interface{}) Op {
	var removedDots []crdt.Dot
	r.elements.Ascend(func(item btree.Item) bool {
		e := item.(elem)
		if e.id != r.siteID {
			removedDots = append(removedDots, crdt.Dot{e.id, e.counter})
		}
		return true
	})
	c := r.summary.Increment(r.siteID)
	r.elements = btree.New(5)
	r.elements.ReplaceOrInsert(newElem(r.siteID, value, c))
	return Op{crdt.Dot{r.siteID, c}, value, removedDots}
}

func (r *Register) ExecuteOp(op Op) interface{} {
	for i := range op.RemovedDots {
		dot := &op.RemovedDots[i]
		v := r.elements.Get(key(dot.SiteID))
		if v == nil {
			continue
		}
		e := v.(elem)
		if e.counter <= dot.Counter {
			r.elements.Delete(key(dot.SiteID))
		}
	}
	r.summary.Insert(&op.Dot)
	v, ins := r.elements.Get(key(op.Dot.SiteID)), true
	if v != nil {
		e := v.(elem)
		if e.counter > op.Dot.Counter {
			ins = false
		}
	}
	if ins {
		r.elements.ReplaceOrInsert(newElem(op.Dot.SiteID, op.Value, op.Dot.Counter))
	}
	return r.Get()
}

func (r *Register) Replicate(siteID crdt.SiteID) *Register {
	crdt.CheckSiteID(siteID)
	return &Register{siteID, r.elements.Clone(), r.summary.Clone()}
}

func (r *Register) Eq(register *Register) bool {
	if r.elements.Len() != register.elements.Len() {
		return false
	}
	if !r.summary.Eq(register.summary) {
		return false
	}
	ok := true
	r.elements.Ascend(func(item btree.Item) bool {
		e := item.(elem)
		i := register.elements.Get(key(e.id))
		if i == nil || e.value != i.(elem).value || e.counter != i.(elem).counter {
			ok = false
		}
		return ok
	})
	return ok
}

type elem struct {
	id      crdt.SiteID
	value   interface{}
	counter crdt.Counter
}

type key crdt.SiteID

func newElem(id crdt.SiteID, value interface{}, counter crdt.Counter) elem {
	return elem{id, value, counter}
}

func (k key) Less(item btree.Item) bool {
	return k < key((item.(elem)).id)
}

func (e elem) Less(item btree.Item) bool {
	switch x := item.(type) {
	case elem:
		return e.id < x.id
	case key:
		return key(e.id) < x
	}
	panic(item)
}
