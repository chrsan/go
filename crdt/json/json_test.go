package json

import (
	"testing"

	"github.com/chrsan/go/crdt"
	"github.com/stretchr/testify/assert"
)

func TestFromString(t *testing.T) {
	j, err := FromString(1, "")
	assert.NoError(t, err)
	assert.Nil(t, j)
	j, err = FromString(2, "123")
	assert.NoError(t, err)
	assert.NotNil(t, 2)
	v, err := j.Value()
	assert.NoError(t, err)
	assert.EqualValues(t, 123, v)
	j, err = FromString(1, "123.45")
	assert.NoError(t, err)
	assert.NotNil(t, j)
	v, err = j.Value()
	assert.NoError(t, err)
	assert.EqualValues(t, 123.45, v)
	j, err = FromString(1, "true")
	assert.NoError(t, err)
	assert.NotNil(t, j)
	v, err = j.Value()
	assert.NoError(t, err)
	assert.EqualValues(t, true, v)
	j, err = FromString(1, "null")
	assert.NoError(t, err)
	assert.NotNil(t, j)
	v, err = j.Value()
	assert.NoError(t, err)
	assert.Nil(t, v)
	j, err = FromString(1, `"hello"`)
	assert.NoError(t, err)
	assert.NotNil(t, j)
	v, err = j.Value()
	assert.NoError(t, err)
	assert.EqualValues(t, "hello", v)
	j, err = FromString(1, `{"foo": 123}`)
	assert.NoError(t, err)
	assert.NotNil(t, j)
	v, err = j.Value()
	assert.NoError(t, err)
	assert.EqualValues(t, map[string]interface{}{"foo": int64(123)}, v)
	j, err = FromString(1, `{"foo": 123, "bar": true, "baz": [1.0, 2.0, 3.0]}`)
	assert.NoError(t, err)
	assert.NotNil(t, j)
	v, err = j.Value()
	assert.NoError(t, err)
	assert.EqualValues(t, map[string]interface{}{"foo": int64(123), "bar": true, "baz": []interface{}{1.0, 2.0, 3.0}}, v)
}

func TestObjectInsert(t *testing.T) {
	j, err := New(1, map[string]interface{}{})
	assert.NoError(t, err)
	op1, err := j.Insert(map[string]float64{"bar": 3.5}, "foo")
	assert.NoError(t, err)
	op2, err := j.Insert(true, "foo", "baz")
	assert.NoError(t, err)
	assert.EqualValues(t, 3, j.summary.Counter(1))
	v, err := j.Value("foo", "bar")
	assert.NoError(t, err)
	assert.Equal(t, 3.5, v)
	v, err = j.Value("foo", "baz")
	assert.NoError(t, err)
	assert.Equal(t, true, v)
	assert.Nil(t, op1.Pointer)
	assert.IsType(t, InnerObjectOp{}, op1.Inner)
	assert.Equal(t, []UID{ObjectUID{Key: "foo", Dot: crdt.Dot{SiteID: 1, Counter: 2}}}, op2.Pointer)
}

func TestObjectInsertInvalidPointer(t *testing.T) {
	j, _ := FromString(1, "{}")
	_, err := j.Insert(map[string]float64{"bar": 3.5}, "foo", "bar")
	assert.Error(t, err)
}
