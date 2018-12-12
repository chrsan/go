package register

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	r := New(1, 8142)
	assert.Equal(t, 8142, r.Get())
	assert.EqualValues(t, 1, r.SiteID())
	assert.EqualValues(t, 1, r.Counter(1))
}

func TestUpdate(t *testing.T) {
	r := New(1, 8142)
	op := r.Update(42)
	assert.Equal(t, 42, r.Get())
	assert.EqualValues(t, 2, r.Counter(1))
	assert.EqualValues(t, 1, op.Dot.SiteID)
	assert.EqualValues(t, 2, op.Dot.Counter)
	assert.Equal(t, 42, op.Value)
	assert.Nil(t, op.RemovedDots)
}

func TestExecuteOp(t *testing.T) {
	r1 := New(1, "a")
	r2 := r1.Replicate(2)
	op := r1.Update("b")
	assert.Equal(t, "b", r2.ExecuteOp(op))
	assert.True(t, r2.Eq(r1))
}

func TestExecuteOpConcurrent(t *testing.T) {
	r1 := New(1, "a")
	r2 := r1.Replicate(2)
	r3 := r1.Replicate(3)
	op1 := r1.Update("b")
	op2 := r2.Update("c")
	op3 := r3.Update("d")
	assert.Equal(t, "b", r1.ExecuteOp(op2))
	assert.Equal(t, "b", r1.ExecuteOp(op3))
	assert.Equal(t, "c", r2.ExecuteOp(op3))
	assert.Equal(t, "b", r2.ExecuteOp(op1))
	assert.Equal(t, "c", r3.ExecuteOp(op2))
	assert.Equal(t, "b", r3.ExecuteOp(op1))
	assert.True(t, r1.Eq(r2))
	assert.True(t, r1.Eq(r3))
}

func TestRemoteDupe(t *testing.T) {
	r1 := New(1, "a")
	r2 := r1.Replicate(2)
	op := r1.Update("b")
	assert.Equal(t, "b", r2.ExecuteOp(op))
	assert.Equal(t, "b", r2.ExecuteOp(op))
	assert.True(t, r1.Eq(r2))
}

func TestExecuteOps(t *testing.T) {
	r1 := New(1, 1)
	r2 := r1.Replicate(2)
	op1 := r1.Update(2)
	assert.Equal(t, 2, r1.Get())
	assert.Equal(t, 2, r2.ExecuteOp(op1))
	op2 := r2.Update(3)
	assert.Equal(t, 3, r2.Get())
	assert.Equal(t, 3, r1.ExecuteOp(op2))
}
