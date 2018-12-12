package text

import (
	"fmt"

	"github.com/chrsan/go/crdt"
	"github.com/google/btree"
)

type tree struct {
	elements *btree.BTree
	count    int
}

func newTree() *tree {
	return &tree{btree.New(5), 0}
}

func (t *tree) clone() *tree {
	return &tree{t.elements.Clone(), t.count}
}

func (t *tree) len() int {
	return t.count
}

func (t *tree) insert(e elem) bool {
	if t.elements.Has(e) {
		return false
	}
	t.elements.ReplaceOrInsert(e)
	t.count += len(e.text)
	return true
}

func (t *tree) remove(uid *crdt.UID) *elem {
	item := t.elements.Delete(elem{uid: uid})
	if item == nil {
		return nil
	}
	e := item.(elem)
	t.count -= len(e.text)
	return &e
}

func (t *tree) lookup(uid *crdt.UID) *elem {
	item := t.elements.Get(elem{uid: uid})
	if item == nil {
		return nil
	}
	e := item.(elem)
	return &e
}

func (t *tree) getElement(index int) (*elem, int) {
	if index >= t.count {
		panic(fmt.Sprintf("Index out of bounds: %d", index))
	}
	var e elem
	ok := false
	t.elements.Ascend(func(i btree.Item) bool {
		x := i.(elem)
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

func (t *tree) getIndex(uid *crdt.UID) int {
	item := t.elements.Get(elem{uid: uid})
	if item == nil {
		return -1
	}
	index := 0
	t.elements.AscendLessThan(item, func(i btree.Item) bool {
		x := i.(elem)
		index += len(x.text)
		return true
	})
	return index
}

type elem struct {
	uid  *crdt.UID
	text string
}

func (e *elem) element() Element {
	return Element{*e.uid, e.text}
}

func (e elem) Less(item btree.Item) bool {
	return e.uid.Cmp(item.(elem).uid) < 0
}

var (
	startTreeElem = &elem{uid: &crdt.MinUID}
	endTreeElem   = &elem{uid: &crdt.MaxUID}
)

func elementBetween(uid1, uid2 *crdt.UID, text string, dot *crdt.Dot) Element {
	uid := crdt.UIDBetween(uid1, uid2, dot)
	return Element{uid, text}
}
