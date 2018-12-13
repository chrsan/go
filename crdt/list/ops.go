package list

import (
	"fmt"

	"github.com/chrsan/go/crdt"
)

type Op interface {
	InsertedElement() (Element, bool)
	RemovedUID() (crdt.UID, bool)
	Validate(siteID crdt.SiteID) error
	isOp()
}

type InsertOp struct {
	Element Element
}

func (i InsertOp) InsertedElement() (Element, bool) {
	return i.Element, true
}

func (i InsertOp) RemovedUID() (crdt.UID, bool) {
	return crdt.UID{}, false
}

func (i InsertOp) Validate(siteID crdt.SiteID) error {
	if i.Element.UID.Dot.SiteID != siteID {
		return fmt.Errorf("Invalid op: %d != %d", i.Element.UID.Dot.SiteID, siteID)
	}
	return nil
}

func (InsertOp) isOp() {}

type RemoveOp struct {
	UID crdt.UID
}

func (r RemoveOp) InsertedElement() (Element, bool) {
	return Element{}, false
}

func (r RemoveOp) RemovedUID() (crdt.UID, bool) {
	return r.UID, true
}

func (r RemoveOp) Validate(siteID crdt.SiteID) error {
	return nil
}

func (RemoveOp) isOp() {}

type LocalOp interface {
	isLocalOp()
}

type LocalInsertOp struct {
	Index int
	Value interface{}
}

func (LocalInsertOp) isLocalOp() {}

type LocalRemoveOp int

func (LocalRemoveOp) isLocalOp() {}
