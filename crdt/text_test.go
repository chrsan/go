package crdt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewText(t *testing.T) {
	x := NewText(1)
	assert.Equal(t, 0, x.Len())
	assert.Equal(t, "", x.String())
}

func TestTextReplace(t *testing.T) {
	x := NewText(1)
	op1, ok1 := x.Replace(0, 0, "Hěllo Ťhere")
	assert.True(t, ok1)
	op2, ok2 := x.Replace(7, 3, "")
	assert.True(t, ok2)
	op3, ok3 := x.Replace(9, 1, "stwhile")
	assert.True(t, ok3)
	assert.Equal(t, "Hěllo erstwhile", x.String())
	assert.Equal(t, 16, x.Len())
	assert.Equal(t, "Hěllo Ťhere", op1.InsertedElements[0].Text)
	assert.Equal(t, 0, op2.RemovedUIDs[0].Cmp(&op1.InsertedElements[0].UID))
	assert.Equal(t, "Hěllo ere", op2.InsertedElements[0].Text)
	assert.Equal(t, 0, op3.RemovedUIDs[0].Cmp(&op2.InsertedElements[0].UID))
	assert.Equal(t, "Hěllo erstwhile", op3.InsertedElements[0].Text)
}

func TestTextReplaceOutOfBounds(t *testing.T) {
	x := NewText(1)
	_, ok := x.Replace(0, 0, "Hěllo Ťhere")
	assert.True(t, ok)
	assert.Panics(t, func() { x.Replace(15, 2, "") })
}

func TestTextExecuteOp(t *testing.T) {
	x1 := NewText(1)
	x2 := x1.Replicate(2)
	op1, ok1 := x1.Replace(0, 0, "Hěllo Ťhere")
	assert.True(t, ok1)
	op2, ok2 := x1.Replace(7, 3, "")
	assert.True(t, ok2)
	op3, ok3 := x1.Replace(9, 1, "stwhile")
	assert.True(t, ok3)
	e1 := x2.ExecuteOp(op1)
	e2 := x2.ExecuteOp(op2)
	e3 := x2.ExecuteOp(op3)
	assert.True(t, x1.state.eq(x2.state))
	assert.Equal(t, 1, len(e1))
	assert.Equal(t, 1, len(e2))
	assert.Equal(t, 1, len(e3))
	assert.Equal(t, TextEdit{0, 0, "Hěllo Ťhere"}, e1[0])
	assert.Equal(t, TextEdit{0, 13, "Hěllo ere"}, e2[0])
	assert.Equal(t, TextEdit{0, 10, "Hěllo erstwhile"}, e3[0])
	e1[0].tryMerge(e2[0].Index, e2[0].Count, e2[0].Text)
	e1[0].tryMerge(e3[0].Index, e3[0].Count, e3[0].Text)
	assert.Equal(t, TextEdit{0, 0, "Hěllo erstwhile"}, e1[0])
}

func TestTextExecuteOpDupe(t *testing.T) {
	x1 := NewText(1)
	x2 := x1.Replicate(2)
	op, ok := x1.Replace(0, 0, "Hiya")
	assert.True(t, ok)
	e1 := x2.ExecuteOp(op)
	e2 := x2.ExecuteOp(op)
	assert.True(t, x1.state.eq(x2.state))
	assert.Equal(t, 1, len(e1))
	assert.Equal(t, 0, len(e2))
	assert.Equal(t, x1.String(), x2.String())
}
