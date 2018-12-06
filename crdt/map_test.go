package crdt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMap(t *testing.T) {
	m := NewMap(1)
	assert.Equal(t, SiteID(1), m.SiteID())
	assert.False(t, m.Contains(412))
	assert.Equal(t, Counter(0), m.summary.Counter(1))
}

func TestMapContains(t *testing.T) {
	m := NewMap(1)
	assert.False(t, m.Contains(123))
	m.Insert(123, -123)
	assert.True(t, m.Contains(123))
}

func TestMapInsert(t *testing.T) {
	m := NewMap(1)
	m.Insert("a", "x")
	op := m.Insert("a", "y")
	assert.NotNil(t, op.InsertedElement)
	assert.Equal(t, "y", op.InsertedElement.Value)
	assert.Equal(t, Dot{1, 2}, op.InsertedElement.Dot)
	assert.Equal(t, []Dot{Dot{1, 1}}, op.RemovedDots)
	assert.Equal(t, 1, m.Len())
}

func TestMapRemove(t *testing.T) {
	m := NewMap(1)
	m.Insert(3, true)
	op, ok := m.Remove(3)
	assert.True(t, ok)
	assert.Equal(t, 3, op.Key)
	assert.Nil(t, op.InsertedElement)
	assert.Equal(t, []Dot{Dot{1, 1}}, op.RemovedDots)
	assert.Equal(t, 0, m.Len())
}

func TestMapRemoveDoesNotExist(t *testing.T) {
	m := NewMap(1)
	_, ok := m.Remove(3)
	assert.False(t, ok)
}

func TestMapExecuteOpInsert(t *testing.T) {
	m1 := NewMap(1)
	m2 := m1.Replicate(2)
	op := m1.Insert(123, 1010)
	lop, ok := m2.ExecuteOp(op)
	assert.True(t, ok)
	assert.Equal(t, LocalMapOp{true, 123, 1010}, lop)
	assert.Equal(t, 1010, m2.Get(123))
	assert.Equal(t, m1.Len(), m2.Len())
	assert.True(t, m1.Eq(m2))
}

func TestMapExecuteOpInsertConcurrent(t *testing.T) {
	m1 := NewMap(1)
	m2 := m1.Replicate(2)
	op1 := m1.Insert(123, 2222)
	op2 := m2.Insert(123, 1111)
	lop1, ok1 := m1.ExecuteOp(op2)
	assert.True(t, ok1)
	assert.Equal(t, LocalMapOp{true, 123, 2222}, lop1)
	lop2, ok2 := m2.ExecuteOp(op1)
	assert.True(t, ok2)
	assert.Equal(t, LocalMapOp{true, 123, 2222}, lop2)
	assert.Equal(t, 2222, m1.Get(123))
	assert.True(t, m1.Eq(m2))
}

func TestMapExecuteOpInsertDupe(t *testing.T) {
	m1 := NewMap(1)
	m2 := m1.Replicate(2)
	op := m1.Insert(123, 2222)
	m2.ExecuteOp(op)
	m3 := m2.Replicate(3)
	lop, ok := m2.ExecuteOp(op)
	assert.True(t, ok)
	assert.Equal(t, LocalMapOp{true, 123, 2222}, lop)
	assert.True(t, m2.Eq(m3))
	assert.True(t, m1.Eq(m2))
}

func TestMapExecuteOpRemove(t *testing.T) {
	m1 := NewMap(1)
	m2 := m1.Replicate(2)
	op1 := m1.Insert(123, 2222)
	op2, ok1 := m1.Remove(123)
	assert.True(t, ok1)
	m2.ExecuteOp(op1)
	assert.Equal(t, 1, m2.Len())
	lop, ok2 := m2.ExecuteOp(op2)
	assert.True(t, ok2)
	assert.Equal(t, LocalMapOp{false, 123, nil}, lop)
	assert.True(t, m2.Eq(m1))
}

func TestMapExecuteOpRemoveDoesNotExist(t *testing.T) {

}
