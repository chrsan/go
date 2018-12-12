package json

import (
	"github.com/chrsan/go/crdt"
	"github.com/chrsan/go/crdt/list"
	"github.com/chrsan/go/crdt/ormap"
	"github.com/chrsan/go/crdt/text"
)

type Op struct {
	Pointer []UID
	Inner   InnerOp
}

func (o Op) InsertedDots() []crdt.Dot {
	switch i := o.Inner.(type) {
	case InnerObjectOp:
		if i.InsertedElement == nil {
			return nil
		}
		return []crdt.Dot{i.InsertedElement.Dot}
	case InnerArrayOp:
		return i.Op.InsertedDots()
	case InnerStringOp:
		var dots []crdt.Dot
		text.Op(i).InsertedDots(func(d *crdt.Dot) {
			dots = append(dots, *d)
		})
		return dots
	}
	panic("")
}

func (o Op) Validate(siteID crdt.SiteID) error {
	switch i := o.Inner.(type) {
	case InnerObjectOp:
		return ormap.Op(i).Validate(siteID)
	case InnerArrayOp:
		return i.Op.Validate(siteID)
	case InnerStringOp:
		return text.Op(i).Validate(siteID)
	}
	panic("")
}

type LocalOp interface {
	Pointer() []LocalUID
	isLocalOp()
}

type LocalInsertOp struct {
	Value   interface{}
	pointer []LocalUID
}

func (i LocalInsertOp) Pointer() []LocalUID {
	return i.pointer
}

func (LocalInsertOp) isLocalOp() {}

type LocalRemoveOp []LocalUID

func (r LocalRemoveOp) Pointer() []LocalUID {
	return r
}

func (LocalRemoveOp) isLocalOp() {}

type LocalReplaceTextOp struct {
	Changes []text.Edit
	pointer []LocalUID
}

func (t LocalReplaceTextOp) Pointer() []LocalUID {
	return t.pointer
}

func (LocalReplaceTextOp) isLocalOp() {}

type UID interface {
	isUID()
}

type ObjectUID struct {
	Key string
	Dot crdt.Dot
}

func (ObjectUID) isUID() {}

type ArrayUID crdt.UID

func (ArrayUID) isUID() {}

type LocalUID interface {
	isLocalUID()
}

type LocalObjectUID string

func (LocalObjectUID) isLocalUID() {}

type LocalArrayUID int

func (LocalArrayUID) isLocalUID() {}

type InnerOp interface {
	isInnerOp()
}

type InnerObjectOp ormap.Op

func (InnerObjectOp) isInnerOp() {}

type InnerArrayOp struct {
	Op list.Op
}

func (InnerArrayOp) isInnerOp() {}

type InnerStringOp text.Op

func (InnerStringOp) isInnerOp() {}
