package text

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chrsan/go/crdt"
)

type Op struct {
	InsertedElements []Element
	RemovedUIDs      []crdt.UID
}

func (o Op) Validate(siteID crdt.SiteID) error {
	var invalid []string
	for i := range o.InsertedElements {
		e := &o.InsertedElements[i]
		if e.UID.Dot.SiteID != siteID {
			invalid = append(invalid, strconv.FormatUint(uint64(e.UID.Dot.SiteID), 10))
		}
	}
	if len(invalid) != 0 {
		return fmt.Errorf("Invalid op: %s != %d", strings.Join(invalid, ", "), siteID)
	}
	return nil
}
