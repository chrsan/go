package ormap

import (
	"fmt"

	"github.com/chrsan/go/crdt"
)

type Op struct {
	Key             interface{}
	InsertedElement *Element
	RemovedDots     []crdt.Dot
}

func (o Op) Validate(siteID crdt.SiteID) error {
	if o.InsertedElement != nil && o.InsertedElement.Dot.SiteID != siteID {
		return fmt.Errorf("Invalid op: %d != %d", o.InsertedElement.Dot.SiteID, siteID)
	}
	return nil
}

type LocalOp struct {
	IsInsert   bool
	Key, Value interface{}
}
