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

func (o Op) Validate(siteID crdt.SiteID) error {
	switch inner := o.Inner.(type) {
	case InnerObjectOp:
		return ormap.Op(inner).Validate(siteID)
	case InnerArrayOp:
		return inner.Op.Validate(siteID)
	case InnerStringOp:
		return text.Op(inner).Validate(siteID)
	default:
		panic(inner)
	}
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
