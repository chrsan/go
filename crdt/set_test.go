package crdt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSet(t *testing.T) {
	s := NewSet(1)
	assert.Equal(t, SiteID(1), s.SiteID())
	assert.False(t, s.Contains(41))
	assert.Equal(t, Counter(0), s.summary.Counter(1))
}

func TestSetContains(t *testing.T) {
	s := NewSet(1)
	assert.Equal(t, SiteID(1), s.SiteID())
	assert.False(t, s.Contains(41))
	s.Insert(41)
	assert.True(t, s.Contains(41))
}

func TestSetInsert(t *testing.T) {
	s := NewSet(1)
	op := s.Insert(123)
	assert.Equal(t, 123, op.Value)
	assert.Nil(t, op.RemovedDots)
	assert.NotNil(t, op.InsertedDot)
	assert.Equal(t, Dot{1, 1}, *op.InsertedDot)
}

func TestSetInsertAlreadyExists(t *testing.T) {
	s := NewSet(1)
	op1 := s.Insert(123)
	op2 := s.Insert(123)
	assert.Equal(t, 1, len(op2.RemovedDots))
	assert.Equal(t, *op1.InsertedDot, op2.RemovedDots[0])
	assert.Equal(t, []interface{}{123}, s.Values())
}

func TestSetRemove(t *testing.T) {
	s := NewSet(1)
	op1 := s.Insert(123)
	op2, ok := s.Remove(123)
	assert.Equal(t, 123, op1.Value)
	assert.NotNil(t, op1.InsertedDot)
	assert.Equal(t, Dot{1, 1}, *op1.InsertedDot)
	assert.Nil(t, op1.RemovedDots)
	assert.True(t, ok)
	assert.Equal(t, 123, op2.Value)
	assert.Nil(t, op2.InsertedDot)
	assert.Equal(t, 1, len(op2.RemovedDots))
	assert.Equal(t, Dot{1, 1}, op2.RemovedDots[0])
	assert.Equal(t, 0, s.Len())
	assert.Equal(t, []interface{}{}, s.Values())
}

func TestSetRemoveDoesNotExist(t *testing.T) {
	s := NewSet(1)
	_, ok := s.Remove(123)
	assert.False(t, ok)
}

func TestSetRemoteInsert(t *testing.T) {
	s1 := NewSet(1)
	s2 := s1.Replicate(8403)
	op := s1.Insert(22)
	lop, ok := s2.ExecuteOp(op)
	assert.True(t, ok)
	assert.Equal(t, LocalSetOp{SetOpInsert, 22}, lop)
}

func TestSetRemoteInsertValueAlreadyExists(t *testing.T) {
	s1 := NewSet(1)
	s2 := s1.Replicate(2)
	op1 := s1.Insert(22)
	op2, ok1 := s1.Remove(22)
	s2.Insert(22)
	assert.True(t, ok1)
	_, ok2 := s2.ExecuteOp(op1)
	assert.False(t, ok2)
	_, ok3 := s2.ExecuteOp(op2)
	assert.False(t, ok3)
	assert.True(t, s2.Contains(22))
}

func TestSetRemoteInsertDupe(t *testing.T) {
	s1 := NewSet(1)
	s2 := s1.Replicate(2)
	op := s1.Insert(22)
	s2.ExecuteOp(op)
	_, ok := s2.ExecuteOp(op)
	assert.False(t, ok)
	assert.Equal(t, s1.Len(), s2.Len())
}

func TestSetExecuteRemoteRemove(t *testing.T) {
	s1 := NewSet(1)
	s2 := s1.Replicate(2)
	s1.Insert(10)
	op, ok1 := s1.Remove(10)
	assert.True(t, ok1)
	assert.Equal(t, 0, s1.Len())
	_, ok2 := s2.ExecuteOp(op)
	assert.False(t, ok2)
	assert.False(t, s2.Contains(10))
	assert.Equal(t, s1.Len(), s2.Len())
}

func TestSetExecuteRemoveSomeDotsRemain(t *testing.T) {
	s1 := NewSet(1)
	s2 := s1.Replicate(2)
	s1.Insert(10)
	s2.Insert(10)
	op, ok1 := s1.Remove(10)
	assert.True(t, ok1)
	_, ok2 := s2.ExecuteOp(op)
	assert.False(t, ok2)
	assert.True(t, s2.Contains(10))
}

func TestSetExecuteRemoteRemoveDupe(t *testing.T) {
	s1 := NewSet(1)
	s2 := s1.Replicate(2)
	op1 := s1.Insert(10)
	op2, ok1 := s1.Remove(10)
	assert.True(t, ok1)
	lop1, ok2 := s2.ExecuteOp(op1)
	assert.True(t, ok2)
	assert.Equal(t, LocalSetOp{SetOpInsert, 10}, lop1)
	lop2, ok3 := s2.ExecuteOp(op2)
	assert.True(t, ok3)
	assert.Equal(t, LocalSetOp{SetOpRemove, 10}, lop2)
	_, ok4 := s2.ExecuteOp(op2)
	assert.False(t, ok4)
	assert.False(t, s2.Contains(10))
}
