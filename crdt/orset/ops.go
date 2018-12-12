package orset

import (
	"fmt"

	"github.com/chrsan/go/crdt"
)

type OpKind int

const (
	OpInsert OpKind = iota
	OpRemove
)

type Op struct {
	Kind        OpKind
	Value       interface{}
	InsertedDot *crdt.Dot
	RemovedDots []crdt.Dot
}

func (o Op) Validate(siteID crdt.SiteID) error {
	if o.InsertedDot != nil && o.InsertedDot.SiteID != siteID {
		return fmt.Errorf("Invalid op: %d != %d", o.InsertedDot.SiteID, siteID)
	}
	return nil
}

type LocalOp struct {
	Kind  OpKind
	Value interface{}
}
