package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TesNew(t *testing.T) {
	l := New(1)
	assert.Equal(t, 0, l.Len())
}

func TestGet(t *testing.T) {
	l := New(1)
	l.Insert(0, 123)
	assert.Equal(t, 123, l.Get(0))
}

func TestInsertPrepend(t *testing.T) {
	l := New(1)
	op1 := l.Insert(0, 123)
	op2 := l.Insert(0, 456)
	op3 := l.Insert(0, 789)
	assert.Equal(t, 3, l.Len())
	assert.Equal(t, 789, l.Get(0))
	assert.Equal(t, 456, l.Get(1))
	assert.Equal(t, 123, l.Get(2))
	e1, ok1 := op1.InsertedElement()
	e2, ok2 := op2.InsertedElement()
	e3, ok3 := op3.InsertedElement()
	assert.True(t, ok1)
	assert.True(t, ok2)
	assert.True(t, ok3)
	assert.True(t, e1.UID.Cmp(&e2.UID) > 0)
	assert.True(t, e2.UID.Cmp(&e3.UID) > 0)
}

func TestInsertAppend(t *testing.T) {
	l := New(1)
	op1 := l.Insert(0, 123)
	op2 := l.Insert(1, 456)
	op3 := l.Insert(2, 789)
	assert.Equal(t, 3, l.Len())
	assert.Equal(t, 123, l.Get(0))
	assert.Equal(t, 456, l.Get(1))
	assert.Equal(t, 789, l.Get(2))
	e1, ok1 := op1.InsertedElement()
	e2, ok2 := op2.InsertedElement()
	e3, ok3 := op3.InsertedElement()
	assert.True(t, ok1)
	assert.True(t, ok2)
	assert.True(t, ok3)
	assert.True(t, e1.UID.Cmp(&e2.UID) < 0)
	assert.True(t, e2.UID.Cmp(&e3.UID) < 0)
}

func TestInsertMiddle(t *testing.T) {
	l := New(1)
	op1 := l.Insert(0, 123)
	op2 := l.Insert(1, 456)
	op3 := l.Insert(1, 789)
	assert.Equal(t, 3, l.Len())
	assert.Equal(t, 123, l.Get(0))
	assert.Equal(t, 789, l.Get(1))
	assert.Equal(t, 456, l.Get(2))
	e1, ok1 := op1.InsertedElement()
	e2, ok2 := op2.InsertedElement()
	e3, ok3 := op3.InsertedElement()
	assert.True(t, ok1)
	assert.True(t, ok2)
	assert.True(t, ok3)
	assert.True(t, e1.UID.Cmp(&e2.UID) < 0)
	assert.True(t, e3.UID.Cmp(&e2.UID) < 0)
}

func TestInsertOutOfBounds(t *testing.T) {
	l := New(1)
	assert.Panics(t, func() { l.Insert(1, 123) })
}

func TestRemove(t *testing.T) {
	l := New(1)
	l.Push(123)
	op1 := l.Push(456)
	l.Push(789)
	_, op2 := l.Remove(1)
	assert.Equal(t, 2, l.Len())
	assert.Equal(t, 123, l.Get(0))
	assert.Equal(t, 789, l.Get(1))
	e, ok := op1.InsertedElement()
	assert.True(t, ok)
	u, ok := op2.RemovedUID()
	assert.True(t, ok)
	assert.True(t, e.UID.Cmp(&u) == 0)
}

func TestRemoveOutOfBounds(t *testing.T) {
	l := New(1)
	assert.Panics(t, func() { l.Remove(0) })
}

func TestPop(t *testing.T) {
	l := New(1)
	op1 := l.Push(123)
	_, op2 := l.Pop()
	assert.Equal(t, 0, l.Len())
	e, ok := op1.InsertedElement()
	assert.True(t, ok)
	assert.Equal(t, 123, e.Value)
	u, ok := op2.RemovedUID()
	assert.True(t, ok)
	assert.True(t, e.UID.Cmp(&u) == 0)
}

func TestPopOutOfBounds(t *testing.T) {
	l := New(1)
	assert.Panics(t, func() { l.Pop() })
}

func TestExecuteOpInsert(t *testing.T) {
	l1 := New(1)
	l2 := l1.Replicate(2)
	op := l1.Push("a")
	lop, ok := l2.ExecuteOp(op)
	assert.Equal(t, 1, l2.Len())
	assert.Equal(t, "a", l2.Get(0))
	assert.True(t, ok)
	assert.Equal(t, LocalInsertOp{Index: 0, Value: "a"}, lop)
	assert.Equal(t, listValues(l1), listValues(l2))
}

func TestExecuteOpInsertDupe(t *testing.T) {
	l1 := New(1)
	l2 := l1.Replicate(2)
	op := l1.Insert(0, "a")
	lop1, ok1 := l2.ExecuteOp(op)
	_, ok2 := l2.ExecuteOp(op)
	assert.Equal(t, 1, l2.Len())
	assert.Equal(t, "a", l2.Get(0))
	assert.True(t, ok1)
	assert.Equal(t, LocalInsertOp{Index: 0, Value: "a"}, lop1)
	assert.False(t, ok2)
	assert.Equal(t, listValues(l1), listValues(l2))
}

func TestExecuteOpRemove(t *testing.T) {
	l1 := New(1)
	l2 := l1.Replicate(2)
	op1 := l1.Push("a")
	_, op2 := l1.Pop()
	lop1, ok1 := l2.ExecuteOp(op1)
	lop2, ok2 := l2.ExecuteOp(op2)
	assert.Equal(t, 0, l1.Len())
	assert.True(t, ok1)
	assert.Equal(t, LocalInsertOp{Index: 0, Value: "a"}, lop1)
	assert.True(t, ok2)
	assert.Equal(t, LocalRemoveOp(0), lop2)
	assert.Equal(t, 0, l2.Len())
}

func TestExecuteRemoveOpDupe(t *testing.T) {
	l1 := New(1)
	l2 := l1.Replicate(2)
	op1 := l1.Push("a")
	_, op2 := l1.Pop()
	lop1, ok1 := l2.ExecuteOp(op1)
	lop2, ok2 := l2.ExecuteOp(op2)
	_, ok3 := l2.ExecuteOp(op2)
	assert.Equal(t, 0, l1.Len())
	assert.True(t, ok1)
	assert.Equal(t, LocalInsertOp{Index: 0, Value: "a"}, lop1)
	assert.True(t, ok2)
	assert.Equal(t, LocalRemoveOp(0), lop2)
	assert.False(t, ok3)
	assert.Equal(t, 0, l2.Len())
}

func TestExecuteOps(t *testing.T) {
	l1 := New(1)
	l2 := New(2)
	op1 := l1.Push(2)
	op2 := l2.Push(1)
	assert.Equal(t, []interface{}{2}, listValues(l1))
	assert.Equal(t, []interface{}{1}, listValues(l2))
	l2.ExecuteOp(op1)
	l1.ExecuteOp(op2)
	assert.Equal(t, listValues(l1), listValues(l2))
}

func listValues(l *List) []interface{} {
	var vs []interface{}
	l.Values(func(v interface{}) {
		vs = append(vs, v)
	})
	return vs
}
