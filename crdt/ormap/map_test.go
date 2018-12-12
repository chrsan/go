package ormap

import (
	"testing"

	"github.com/chrsan/go/crdt"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	m := New(1)
	assert.EqualValues(t, 1, m.SiteID())
	assert.False(t, m.Contains(412))
	assert.EqualValues(t, 0, m.summary.Counter(1))
}

func TestContains(t *testing.T) {
	m := New(1)
	assert.False(t, m.Contains(123))
	m.Insert(123, -123)
	assert.True(t, m.Contains(123))
}

func TestInsert(t *testing.T) {
	m := New(1)
	m.Insert("a", "x")
	op := m.Insert("a", "y")
	assert.NotNil(t, op.InsertedElement)
	assert.Equal(t, "y", op.InsertedElement.Value)
	assert.Equal(t, crdt.Dot{1, 2}, op.InsertedElement.Dot)
	assert.Equal(t, []crdt.Dot{crdt.Dot{1, 1}}, op.RemovedDots)
	assert.Equal(t, 1, m.Len())
}

func TestRemove(t *testing.T) {
	m := New(1)
	m.Insert(3, true)
	op, ok := m.Remove(3)
	assert.True(t, ok)
	assert.Equal(t, 3, op.Key)
	assert.Nil(t, op.InsertedElement)
	assert.Equal(t, []crdt.Dot{crdt.Dot{1, 1}}, op.RemovedDots)
	assert.Equal(t, 0, m.Len())
}

func TestRemoveDoesNotExist(t *testing.T) {
	m := New(1)
	_, ok := m.Remove(3)
	assert.False(t, ok)
}

func TestExecuteOpInsert(t *testing.T) {
	m1 := New(1)
	m2 := m1.Replicate(2)
	op := m1.Insert(123, 1010)
	lop := m2.ExecuteOp(op)
	assert.Equal(t, LocalOp{true, 123, 1010}, lop)
	assert.Equal(t, 1010, m2.Get(123))
	assert.Equal(t, m1.Len(), m2.Len())
	assert.True(t, m1.Eq(m2))
}

func TestExecuteOpInsertConcurrent(t *testing.T) {
	m1 := New(1)
	m2 := m1.Replicate(2)
	op1 := m1.Insert(123, 2222)
	op2 := m2.Insert(123, 1111)
	lop1 := m1.ExecuteOp(op2)
	assert.Equal(t, LocalOp{true, 123, 2222}, lop1)
	lop2 := m2.ExecuteOp(op1)
	assert.Equal(t, LocalOp{true, 123, 2222}, lop2)
	assert.Equal(t, 2222, m1.Get(123))
	assert.True(t, m1.Eq(m2))
}

func TestExecuteOpInsertDupe(t *testing.T) {
	m1 := New(1)
	m2 := m1.Replicate(2)
	op := m1.Insert(123, 2222)
	m2.ExecuteOp(op)
	lop := m2.ExecuteOp(op)
	assert.Equal(t, LocalOp{true, 123, 2222}, lop)
	assert.True(t, m1.Eq(m2))
}

func TestExecuteOpRemove(t *testing.T) {
	m1 := New(1)
	m2 := m1.Replicate(2)
	op1 := m1.Insert(123, 2222)
	op2, ok := m1.Remove(123)
	assert.True(t, ok)
	m2.ExecuteOp(op1)
	assert.Equal(t, 1, m2.Len())
	lop := m2.ExecuteOp(op2)
	assert.Equal(t, LocalOp{false, 123, nil}, lop)
	assert.True(t, m2.Eq(m1))
}

func TestExecuteOpRemoveDoesNotExist(t *testing.T) {
	m1 := New(1)
	m2 := m1.Replicate(2)
	m1.Insert(123, 2222)
	op, ok1 := m1.Remove(123)
	assert.True(t, ok1)
	state := m2.state.Clone()
	lop := m2.ExecuteOp(op)
	assert.Equal(t, LocalOp{false, 123, nil}, lop)
	assert.Equal(t, 0, m2.Len())
	assert.True(t, m2.state.Eq(state))
}

func TestExecuteOpRemoveSomeDotsRemain(t *testing.T) {
	m1 := New(1)
	m2 := m1.Replicate(2)
	m3 := m1.Replicate(3)
	op1 := m2.Insert(123, 1111)
	op2 := m1.Insert(123, 2222)
	op3, ok := m1.Remove(123)
	assert.True(t, ok)
	lop1 := m3.ExecuteOp(op1)
	assert.Equal(t, LocalOp{true, 123, 1111}, lop1)
	lop2 := m3.ExecuteOp(op2)
	assert.Equal(t, LocalOp{true, 123, 2222}, lop2)
	lop3 := m3.ExecuteOp(op3)
	assert.Equal(t, LocalOp{true, 123, 1111}, lop3)
}

func TestExecuteOpRemoveDupe(t *testing.T) {
	m1 := New(1)
	m2 := m1.Replicate(2)
	op1 := m1.Insert(123, 2222)
	op2, ok := m1.Remove(123)
	assert.True(t, ok)
	lop1 := m2.ExecuteOp(op1)
	assert.Equal(t, LocalOp{true, 123, 2222}, lop1)
	lop2 := m2.ExecuteOp(op2)
	assert.Equal(t, LocalOp{false, 123, nil}, lop2)
	lop3 := m2.ExecuteOp(op2)
	assert.Equal(t, LocalOp{false, 123, nil}, lop3)
	assert.Equal(t, 0, m1.Len())
	assert.Equal(t, 0, m2.Len())
}
