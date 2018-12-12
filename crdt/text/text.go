package text

import "github.com/chrsan/go/crdt"

type Text struct {
	siteID  crdt.SiteID
	state   *State
	summary *crdt.Summary
}

func New(siteID crdt.SiteID) *Text {
	crdt.CheckSiteID(siteID)
	return &Text{siteID, NewState(), crdt.NewSummary()}
}

func (t *Text) SiteID() crdt.SiteID {
	return t.siteID
}

func (t *Text) Len() int {
	return t.state.tree.count
}

func (t *Text) Replace(index, count int, text string) (Op, bool) {
	dot := t.summary.Dot(t.siteID)
	return t.state.Replace(index, count, text, &dot)
}

func (t *Text) ExecuteOp(op Op) Edits {
	op.InsertedDots(func(dot *crdt.Dot) {
		t.summary.Insert(dot)
	})
	return t.state.ExecuteOp(op)
}

func (t *Text) String() string {
	return t.state.String()
}

func (t *Text) Replicate(siteID crdt.SiteID) *Text {
	crdt.CheckSiteID(siteID)
	return &Text{siteID, t.state.Clone(), t.summary.Clone()}
}

func (t *Text) Eq(text *Text) bool {
	return t.state.Eq(text.state) && t.summary.Eq(text.summary)
}

type Element struct {
	UID  crdt.UID
	Text string
}
