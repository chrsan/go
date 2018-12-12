package text

import "strings"

type Edits []Edit

type Edit struct {
	Index, Count int
	Text         string
}

func pushEdit(edits Edits, index, count int, text string) Edits {
	if len(edits) == 0 || !edits[len(edits)-1].tryMerge(index, count, text) {
		return append(edits, Edit{index, count, text})
	}
	return edits
}

/*
func compactEdits(edits []*Edit) []*Edit {
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

func (e *Edit) tryOverwrite(index, count int, text string) bool {
	if e.shouldOverwrite(index, count) {
		e.modify(true, index, count, text)
		return true
	}
	return false
}

func (e *Edit) tryMerge(index, count int, text string) bool {
	if e.canMerge(index, count) {
		e.modify(false, index, count, text)
		return true
	}
	return false
}

func (e *Edit) shiftOrDestroy(index, count int, text string) *Edit {
	if index+count <= e.Index {
		e.Index -= count
		e.Index += len(text)
		return e
	}
	if index >= e.Index+len(e.Text) {
		return e
	}
	return nil
}

func (e *Edit) shouldOverwrite(index, count int) bool {
	return e.canMerge(index, count) && len(e.Text) < 64 && !strings.HasSuffix(e.Text, "\n")
}

func (e *Edit) canMerge(index, count int) bool {
	return index+count >= e.Index && index <= e.Index+len(e.Text)
}

func (e *Edit) modify(overwrite bool, index, count int, text string) {
	deletesBefore := saturatingSub(e.Index, index)
	insertIndex := saturatingSub(index, e.Index)
	deletesAfter := count - deletesBefore
	textDeleteLen := min(deletesAfter, len(e.Text)-insertIndex)
	deletesAfter = saturatingSub(deletesAfter, textDeleteLen)
	e.Index = min(e.Index, index)
	if overwrite {
		e.Count = deletesBefore + len(e.Text) + deletesAfter
	} else {
		e.Count += deletesBefore + deletesAfter
	}
	var b strings.Builder
	b.WriteString(e.Text[:insertIndex])
	b.WriteString(text)
	b.WriteString(e.Text[insertIndex+textDeleteLen:])
	e.Text = b.String()
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
