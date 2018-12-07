package crdt

import (
	"fmt"
	"strings"

	"github.com/google/btree"
)

type Text struct {
	siteID  SiteID
	state   *textState
	summary *Summary
}

func NewText(siteID SiteID) *Text {
	checkSiteID(siteID)
	return &Text{siteID, &textState{tree: newTextTree()}, NewSummary()}
}

func (t *Text) SiteID() SiteID {
	return t.siteID
}

func (t *Text) Len() int {
	return t.state.tree.count
}

func (t *Text) Replace(index, count int, text string) (TextOp, bool) {
	dot := t.summary.Dot(t.siteID)
	return t.state.replace(index, count, text, &dot)
}

func (t *Text) ExecuteOp(op TextOp) TextEdits {
	op.InsertedDots(func(dot *Dot) {
		t.summary.Insert(dot)
	})
	return t.state.executeOp(op)
}

func (t *Text) Value() string {
	return t.state.value()
}

func (t *Text) Replicate(siteID SiteID) *Text {
	checkSiteID(siteID)
	return &Text{siteID, t.state.clone(), t.summary.Clone()}
}

func (t *Text) Eq(text *Text) bool {
	return t.state.eq(text.state) && t.summary.Eq(text.summary)
}

type TextElement struct {
	UID  UID
	Text string
}

type TextOp struct {
	InsertedElements []TextElement
	RemovedUIDs      []UID
}

func (t TextOp) InsertedDots(f func(*Dot)) {
	for i := range t.RemovedUIDs {
		uid := &t.RemovedUIDs[i]
		f(&uid.Dot)
	}
}

type textState struct {
	tree *textTree
	edit *TextEdit
}

func (t *textState) replace(index, count int, text string, dot *Dot) (TextOp, bool) {
	if index+count > t.tree.count {
		panic(fmt.Sprintf("Index out of bounds: %d", index))
	}
	if count == 0 && len(text) == 0 {
		return TextOp{}, false
	}
	e := t.genMergedEdit(index, count, text)
	o := t.getElementOffset(e.Index)
	if o == 0 && e.Count == 0 {
		return t.doInsert(e.Index, e.Text, dot), true
	}
	return t.doReplace(e.Index, e.Count, e.Text, dot), true
}

func (t *textState) doInsert(index int, text string, dot *Dot) TextOp {
	uid1, uid2 := t.getPrevUID(index), t.getUID(index)
	elem := textElementBetween(uid1, uid2, text, dot)
	t.tree.insert(textTreeElem{&elem.UID, elem.Text})
	return TextOp{InsertedElements: []TextElement{elem}}
}

func (t *textState) doReplace(index, count int, text string, dot *Dot) TextOp {
	elem, offset := t.removeAt(index)
	i, l := index-offset, len(elem.text)-offset
	removes := []*textTreeElem{elem}
	for l < count {
		e, _ := t.removeAt(i)
		l -= len(e.text)
		removes = append(removes, e)
	}
	var inserts []TextElement
	if offset > 0 || len(text) != 0 || l > count {
		uid1, uid2 := t.getPrevUID(i), t.getUID(i)
		if offset > 0 {
			txt := removes[0].text[:offset]
			inserts = append(inserts, textElementBetween(uid1, uid2, txt, dot))
		}
		if len(text) != 0 {
			u := uid1
			if len(inserts) != 0 {
				u = &inserts[len(inserts)-1].UID
			}
			inserts = append(inserts, textElementBetween(u, uid2, text, dot))
		}
		if l > count {
			old := removes[len(removes)-1]
			txt := old.text[len(old.text)+count-l:]
			u := uid1
			if len(inserts) != 0 {
				u = &inserts[len(inserts)-1].UID
			}
			inserts = append(inserts, textElementBetween(u, uid2, txt, dot))
		}
	}
	for j := range inserts {
		e := &inserts[j]
		t.tree.insert(textTreeElem{&e.UID, e.Text})
	}
	removedUIDs := make([]UID, len(removes))
	for j, e := range removes {
		removedUIDs[j] = *e.uid
	}
	return TextOp{inserts, removedUIDs}
}

func (t *textState) executeOp(op TextOp) TextEdits {
	var edits TextEdits
	for i := range op.RemovedUIDs {
		u := &op.RemovedUIDs[i]
		j := t.tree.getIndex(u)
		if j == -1 {
			continue
		}
		e := t.tree.remove(u)
		if e == nil {
			panic(fmt.Sprintf("Invalid state. UID does not exist: %v", *u))
		}
		edits = pushTextEdit(edits, j, len(e.text), "")
	}
	for i := range op.InsertedElements {
		e := &op.InsertedElements[i]
		if ok := t.tree.insert(textTreeElem{&e.UID, e.Text}); !ok {
			continue
		}
		j := t.tree.getIndex(&e.UID)
		if j == -1 {
			panic(fmt.Sprintf("Invalid state. UID does not exist: %v", e.UID))
		}
		edits = pushTextEdit(edits, j, 0, e.Text)
	}
	t.shiftMergedEdit(edits)
	return edits
}

func (t *textState) clone() *textState {
	return &textState{t.tree.clone(), nil}
}

func (t *textState) value() string {
	var b strings.Builder
	t.tree.elements.Ascend(func(item btree.Item) bool {
		b.WriteString(item.(textTreeElem).text)
		return true
	})
	return b.String()
}

func (t *textState) eq(state *textState) bool {
	if t.tree.count != state.tree.count {
		return false
	}
	ok := true
	t.tree.elements.Ascend(func(item btree.Item) bool {
		e := item.(textTreeElem)
		i := state.tree.elements.Get(textTreeElem{uid: e.uid})
		if i == nil {
			ok = false
		} else if e.uid.Cmp(i.(textTreeElem).uid) != 0 {
			ok = false
		}
		return ok
	})
	return ok
}

func (t *textState) removeAt(index int) (*textTreeElem, int) {
	e, i := t.tree.getElement(index)
	t.tree.remove(e.uid)
	return e, i
}

func (t *textState) getUID(index int) *UID {
	if index == t.tree.count {
		return endTextTreeElem.uid
	}
	e, _ := t.tree.getElement(index)
	return e.uid
}

func (t *textState) getPrevUID(index int) *UID {
	if index == 0 {
		return startTextTreeElem.uid
	}
	e, _ := t.tree.getElement(index)
	return e.uid
}

func (t *textState) getElementOffset(index int) int {
	if index == t.tree.count {
		return 0
	}
	_, i := t.tree.getElement(index)
	return i
}

func (t *textState) genMergedEdit(index, count int, text string) *TextEdit {
	if t.edit != nil {
		if t.edit.tryOverwrite(index, count, text) {
			return &TextEdit{t.edit.Index, t.edit.Count, t.edit.Text}
		}
	}
	t.edit = &TextEdit{index, count, text}
	return &TextEdit{index, count, text}
}

func (t *textState) shiftMergedEdit(edits TextEdits) {
	for _, edit := range edits {
		e := t.edit
		t.edit = nil
		if e == nil {
			break
		}
		t.edit = e.shiftOrDestroy(edit.Index, edit.Count, edit.Text)
	}
}

type TextEdit struct {
	Index, Count int
	Text         string
}

type TextEdits []TextEdit

func pushTextEdit(edits TextEdits, index, count int, text string) TextEdits {
	if len(edits) == 0 || !edits[len(edits)-1].tryMerge(index, count, text) {
		return append(edits, TextEdit{index, count, text})
	}
	return edits
}

// TODO: Needed?
/*
func compactTextEdits(edits []*TextEdit) []*TextEdit {
	if len(edits) <= 1 {
		return edits
	}
	ci := 0
	for i := 1; i < len(edits); i++ {
		e := edits[i]
		if !edits[ci].tryMerge(e.Index, e.Count, e.Text) {
			ci++
			f := edits[ci]
			edits[ci] = e
			edits[i] = f
		}
	}
	if ci+1 < len(edits) {
		return edits[:ci+1]
	}
	return edits
}
*/

func (t *TextEdit) tryOverwrite(index, count int, text string) bool {
	if t.shouldOverwrite(index, count) {
		t.modify(true, index, count, text)
		return true
	}
	return false
}

func (t *TextEdit) tryMerge(index, count int, text string) bool {
	if t.canMerge(index, count) {
		t.modify(false, index, count, text)
		return true
	}
	return false
}

func (t *TextEdit) shiftOrDestroy(index, count int, text string) *TextEdit {
	if index+count <= t.Index {
		t.Index -= count
		t.Index += len(text)
		return t
	}
	if index >= t.Index+len(t.Text) {
		return t
	}
	return nil
}

func (t *TextEdit) shouldOverwrite(index, count int) bool {
	return t.canMerge(index, count) && len(t.Text) < 64 && !strings.HasSuffix(t.Text, "\n")
}

func (t *TextEdit) canMerge(index, count int) bool {
	return index+count >= t.Index && index <= t.Index+len(t.Text)
}

func (t *TextEdit) modify(overwrite bool, index, count int, text string) {
	deletesBefore := saturatingSub(t.Index, index)
	insertIndex := saturatingSub(index, t.Index)
	deletesAfter := count - deletesBefore
	textDeleteLen := min(deletesAfter, len(t.Text)-insertIndex)
	deletesAfter = saturatingSub(deletesAfter, textDeleteLen)
	t.Index = min(t.Index, index)
	if overwrite {
		t.Count = deletesBefore + len(t.Text) + deletesAfter
	} else {
		t.Count += deletesBefore + deletesAfter
	}
	var b strings.Builder
	b.WriteString(t.Text[:insertIndex])
	b.WriteString(text)
	b.WriteString(t.Text[insertIndex+textDeleteLen:])
	t.Text = b.String()
}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func saturatingSub(i, j int) int {
	k := i - j
	if k < 0 {
		return 0
	}
	return k
}

type textTree struct {
	elements *btree.BTree
	count    int
}

func newTextTree() *textTree {
	return &textTree{btree.New(5), 0}
}

func (t *textTree) clone() *textTree {
	return &textTree{t.elements.Clone(), t.count}
}

func (t *textTree) len() int {
	return t.count
}

func (t *textTree) insert(e textTreeElem) bool {
	if t.elements.Has(e) {
		return false
	}
	t.elements.ReplaceOrInsert(e)
	t.count += len(e.text)
	return true
}

func (t *textTree) remove(uid *UID) *textTreeElem {
	item := t.elements.Delete(textTreeElem{uid: uid})
	if item == nil {
		return nil
	}
	e := item.(textTreeElem)
	t.count -= len(e.text)
	return &e
}

func (t *textTree) lookup(uid *UID) *textTreeElem {
	item := t.elements.Get(textTreeElem{uid: uid})
	if item == nil {
		return nil
	}
	e := item.(textTreeElem)
	return &e
}

func (t *textTree) getElement(index int) (*textTreeElem, int) {
	if index >= t.count {
		panic(fmt.Sprintf("Index out of bounds: %d", index))
	}
	var e textTreeElem
	ok := false
	t.elements.Ascend(func(i btree.Item) bool {
		x := i.(textTreeElem)
		if index < len(x.text) {
			e = x
			ok = true
			return false
		}
		index -= len(x.text)
		return true
	})
	if !ok {
		panic(fmt.Sprintf("No element found at index: %d", index))
	}
	return &e, index
}

func (t *textTree) getIndex(uid *UID) int {
	item := t.elements.Get(textTreeElem{uid: uid})
	if item == nil {
		return -1
	}
	index := 0
	t.elements.AscendLessThan(item, func(i btree.Item) bool {
		x := i.(textTreeElem)
		index += len(x.text)
		return true
	})
	return index
}

type textTreeElem struct {
	uid  *UID
	text string
}

func (t *textTreeElem) textElement() TextElement {
	return TextElement{*t.uid, t.text}
}

var (
	startTextTreeElem = &textTreeElem{uid: &MinUID}
	endTextTreeElem   = &textTreeElem{uid: &MaxUID}
)

func textElementBetween(uid1, uid2 *UID, text string, dot *Dot) TextElement {
	uid := UIDBetween(uid1, uid2, dot)
	return TextElement{uid, text}
}

func (e textTreeElem) Less(item btree.Item) bool {
	return e.uid.Cmp(item.(textTreeElem).uid) < 0
}
