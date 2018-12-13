package text

import (
	"fmt"
	"strings"

	"github.com/chrsan/go/crdt"
	"github.com/google/btree"
)

type State struct {
	tree *tree
	edit *Edit
}

func NewState() *State {
	return &State{tree: newTree()}
}

func (s *State) Len() int {
	return s.tree.count
}

func (s *State) Replace(index, count int, text string, dot *crdt.Dot) (Op, bool) {
	if index+count > s.tree.count {
		panic(fmt.Sprintf("Index out of bounds: %d", index))
	}
	if count == 0 && len(text) == 0 {
		return Op{}, false
	}
	e := s.genMergedEdit(index, count, text)
	o := s.getElementOffset(e.Index)
	if o == 0 && e.Count == 0 {
		return s.doInsert(e.Index, e.Text, dot), true
	}
	return s.doReplace(e.Index, e.Count, e.Text, dot), true
}

func (s *State) ExecuteOp(op Op) Edits {
	var edits Edits
	for i := range op.RemovedUIDs {
		u := &op.RemovedUIDs[i]
		j := s.tree.getIndex(u)
		if j == -1 {
			continue
		}
		e := s.tree.remove(u)
		if e == nil {
			panic(fmt.Sprintf("Invalid state. UID does not exist: %v", *u))
		}
		edits = pushEdit(edits, j, len(e.text), "")
		e = nil
	}
	for i := range op.InsertedElements {
		e := &op.InsertedElements[i]
		if ok := s.tree.insert(elem{&e.UID, e.Text}); !ok {
			continue
		}
		j := s.tree.getIndex(&e.UID)
		if j == -1 {
			panic(fmt.Sprintf("Invalid state. UID does not exist: %v", e.UID))
		}
		edits = pushEdit(edits, j, 0, e.Text)
	}
	s.shiftMergedEdit(edits)
	return edits
}

func (s *State) Clone() *State {
	return &State{s.tree.clone(), nil}
}

func (s *State) String() string {
	var b strings.Builder
	s.tree.elements.Ascend(func(item btree.Item) bool {
		b.WriteString(item.(elem).text)
		return true
	})
	return b.String()
}

func (s *State) Eq(state *State) bool {
	if s.tree.count != state.tree.count {
		return false
	}
	ok := true
	s.tree.elements.Ascend(func(item btree.Item) bool {
		e := item.(elem)
		i := state.tree.elements.Get(elem{uid: e.uid})
		if i == nil {
			ok = false
		} else if e.uid.Cmp(i.(elem).uid) != 0 {
			ok = false
		}
		return ok
	})
	return ok
}

func (s *State) doInsert(index int, text string, dot *crdt.Dot) Op {
	uid1, uid2 := s.getPrevUID(index), s.getUID(index)
	e := elementBetween(uid1, uid2, text, dot)
	s.tree.insert(elem{&e.UID, e.Text})
	return Op{InsertedElements: []Element{e}}
}

func (s *State) doReplace(index, count int, text string, dot *crdt.Dot) Op {
	e, offset := s.removeAt(index)
	i, l := index-offset, len(e.text)-offset
	removes := []*elem{e}
	for l < count {
		e, _ := s.removeAt(i)
		l -= len(e.text)
		removes = append(removes, e)
	}
	var inserts []Element
	if offset > 0 || len(text) != 0 || l > count {
		uid1, uid2 := s.getPrevUID(i), s.getUID(i)
		if offset > 0 {
			txt := removes[0].text[:offset]
			inserts = append(inserts, elementBetween(uid1, uid2, txt, dot))
		}
		if len(text) != 0 {
			u := uid1
			if len(inserts) != 0 {
				u = &inserts[len(inserts)-1].UID
			}
			inserts = append(inserts, elementBetween(u, uid2, text, dot))
		}
		if l > count {
			old := removes[len(removes)-1]
			txt := old.text[len(old.text)+count-l:]
			u := uid1
			if len(inserts) != 0 {
				u = &inserts[len(inserts)-1].UID
			}
			inserts = append(inserts, elementBetween(u, uid2, txt, dot))
		}
	}
	for j := range inserts {
		e := &inserts[j]
		s.tree.insert(elem{&e.UID, e.Text})
	}
	removedUIDs := make([]crdt.UID, len(removes))
	for j, e := range removes {
		removedUIDs[j] = *e.uid
		e = nil
	}
	removes = nil
	return Op{inserts, removedUIDs}
}

func (s *State) removeAt(index int) (*elem, int) {
	e, i := s.tree.getElement(index)
	s.tree.remove(e.uid)
	return e, i
}

func (s *State) getUID(index int) *crdt.UID {
	if index == s.tree.count {
		return endTreeElem.uid
	}
	e, _ := s.tree.getElement(index)
	return e.uid
}

func (s *State) getPrevUID(index int) *crdt.UID {
	if index == 0 {
		return startTreeElem.uid
	}
	e, _ := s.tree.getElement(index)
	return e.uid
}

func (s *State) getElementOffset(index int) int {
	if index == s.tree.count {
		return 0
	}
	_, i := s.tree.getElement(index)
	return i
}

func (s *State) genMergedEdit(index, count int, text string) *Edit {
	if s.edit != nil {
		if s.edit.tryOverwrite(index, count, text) {
			return &Edit{s.edit.Index, s.edit.Count, s.edit.Text}
		}
	}
	s.edit = &Edit{index, count, text}
	return &Edit{index, count, text}
}

func (s *State) shiftMergedEdit(edits Edits) {
	for _, edit := range edits {
		e := s.edit
		s.edit = nil
		if e == nil {
			break
		}
		s.edit = e.shiftOrDestroy(edit.Index, edit.Count, edit.Text)
	}
}
