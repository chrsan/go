package crdt

import (
	"fmt"

	"github.com/google/btree"
)

type Register struct {
	siteID   SiteID
	elements *btree.BTree
	summary  *Summary
}

func NewRegister(siteID SiteID, value interface{}) *Register {
	checkSiteID(siteID)
	elements := btree.New(5)
	summary := NewSummary()
	counter := Counter(1)
	summary.Insert(&Dot{SiteID: siteID, Counter: counter})
	elements.ReplaceOrInsert(newElem(siteID, &SiteValue{value, counter}))
	return &Register{siteID, elements, summary}
}

func (r *Register) SiteID() SiteID {
	return r.siteID
}

func (r *Register) Counter(siteID SiteID) Counter {
	return r.summary.Counter(siteID)
}

func (r *Register) Get() interface{} {
	e := r.elements.Min().(elem)
	return e.v.Value
}

func (r *Register) Update(value interface{}) RegisterOp {
	var removedDots []Dot
	r.elements.Ascend(func(item btree.Item) bool {
		e := item.(elem)
		if e.id != r.siteID {
			removedDots = append(removedDots, Dot{e.id, e.v.Counter})
		}
		return true
	})
	c := r.summary.Increment(r.siteID)
	r.elements = btree.New(5)
	r.elements.ReplaceOrInsert(newElem(r.siteID, &SiteValue{value, c}))
	return RegisterOp{Dot{r.siteID, c}, value, removedDots}
}

func (r *Register) ExecuteOp(op RegisterOp) interface{} {
	for i := range op.RemovedDots {
		dot := &op.RemovedDots[i]
		v := r.elements.Get(dot.SiteID)
		if v == nil {
			continue
		}
		e := v.(elem)
		if e.v.Counter <= dot.Counter {
			r.elements.Delete(dot.SiteID)
		}
	}
	r.summary.Insert(&op.Dot)
	v, ins := r.elements.Get(op.Dot.SiteID), true
	if v != nil {
		e := v.(elem)
		if e.v.Counter > op.Dot.Counter {
			ins = false
		}
	}
	if ins {
		r.elements.ReplaceOrInsert(elem{op.Dot.SiteID, &SiteValue{op.Value, op.Dot.Counter}})
	}
	return r.Get()
}

func (r *Register) Replicate(siteID SiteID) *Register {
	checkSiteID(siteID)
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
		i := register.elements.Get(e.id)
		if i == nil {
			ok = false
		} else if *e.v != *(i.(elem).v) {
			ok = false
		}
		return ok
	})
	return ok
}

type SiteValue struct {
	Value   interface{}
	Counter Counter
}

type RegisterOp struct {
	Dot         Dot
	Value       interface{}
	RemovedDots []Dot
}

func (r RegisterOp) Validate(siteID SiteID) error {
	if r.Dot.SiteID != siteID {
		return fmt.Errorf("Invalid op: %d != %d", r.Dot.SiteID, siteID)
	}
	return nil
}

type elem struct {
	id SiteID
	v  *SiteValue
}

func newElem(id SiteID, v *SiteValue) elem {
	return elem{id, v}
}

func (i SiteID) Less(item btree.Item) bool {
	return i < (item.(elem)).id
}

func (e elem) Less(item btree.Item) bool {
	switch x := item.(type) {
	case elem:
		return e.id < x.id
	case SiteID:
		return e.id < x
	}
	panic(item)
}
