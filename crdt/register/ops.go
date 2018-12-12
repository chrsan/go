package register

import (
	"fmt"

	"github.com/chrsan/go/crdt"
)

type Op struct {
	Dot         crdt.Dot
	Value       interface{}
	RemovedDots []crdt.Dot
}

func (o Op) Validate(siteID crdt.SiteID) error {
	if o.Dot.SiteID != siteID {
		return fmt.Errorf("Invalid op: %d != %d", o.Dot.SiteID, siteID)
	}
	return nil
}
