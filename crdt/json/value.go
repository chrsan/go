package json

import (
	"errors"
	"fmt"

	"github.com/chrsan/go/crdt"
	"github.com/chrsan/go/crdt/list"
	"github.com/chrsan/go/crdt/ormap"
	"github.com/chrsan/go/crdt/text"
)

type Value interface {
	Value() interface{}
	clone() Value
	isValue()
}

type Builder struct {
	dot    *crdt.Dot
	root   Value
	values []Value
}

func (b *Builder) StartObject() *Builder {
	o := Object{&ormap.State{}}
	if b.root == nil {
		b.root = o
		b.values = append(b.values, o)
		return b
	}
	b.currentArray().v.Push(o, b.dot)
	b.values = append(b.values, o)
	return b
}

func (b *Builder) EndObject() *Builder {
	b.currentObject()
	b.values = b.values[:len(b.values)-1]
	return b
}

func (b *Builder) StartArray() *Builder {
	a := Array{&list.State{}}
	if b.root == nil {
		b.root = a
		b.values = append(b.values, a)
		return b
	}
	b.currentArray().v.Push(a, b.dot)
	b.values = append(b.values, a)
	return b
}

func (b *Builder) EndArray() *Builder {
	b.currentArray()
	b.values = b.values[:len(b.values)-1]
	return b
}

func (b *Builder) StringValue(s string) String {
	t := text.NewState()
	if s != "" {
		t.Replace(0, 0, s, b.dot)
	}
	return String{t}
}

func (b *Builder) Field(name string, value Value) *Builder {
	o := b.currentObject()
	if value == nil {
		o.v.Insert(name, Null{}, *b.dot)
	} else {
		o.v.Insert(name, value, *b.dot)
	}
	return b
}

func (b *Builder) Array(name string, values []Value) *Builder {
	o := b.currentObject()
	if values == nil {
		o.v.Insert(name, Null{}, *b.dot)
		return b
	}
	a := Array{&list.State{}}
	for _, v := range values {
		if v == nil {
			a.v.Push(Null{}, b.dot)
		} else {
			a.v.Push(v, b.dot)
		}
	}
	o.v.Insert(name, a, *b.dot)
	return b
}

func (b *Builder) StringField(name string, value string) *Builder {
	return b.Field(name, b.StringValue(value))
}

func (b *Builder) StringArray(name string, values []string) *Builder {
	if values == nil {
		return b.Field(name, nil)
	}
	vs := make([]Value, len(values))
	for i, v := range values {
		vs[i] = b.StringValue(v)
	}
	return b.Array(name, vs)
}

func (b *Builder) BooleanField(name string, value bool) *Builder {
	return b.Field(name, Boolean(value))
}

func (b *Builder) BooleanArray(name string, values []bool) *Builder {
	if values == nil {
		return b.Field(name, nil)
	}
	vs := make([]Value, len(values))
	for i, v := range values {
		vs[i] = Boolean(v)
	}
	return b.Array(name, vs)
}

func (b *Builder) IntField(name string, value int64) *Builder {
	return b.Field(name, Int(value))
}

func (b *Builder) IntArray(name string, values []int64) *Builder {
	if values == nil {
		return b.Field(name, nil)
	}
	vs := make([]Value, len(values))
	for i, v := range values {
		vs[i] = Int(v)
	}
	return b.Array(name, vs)
}

func (b *Builder) FloatField(name string, value float64) *Builder {
	return b.Field(name, Float(value))
}

func (b *Builder) FloatArray(name string, values []float64) *Builder {
	if values == nil {
		return b.Field(name, nil)
	}
	vs := make([]Value, len(values))
	for i, v := range values {
		vs[i] = Float(v)
	}
	return b.Array(name, vs)
}

func (b *Builder) StartObjectField(name string) *Builder {
	o := Object{&ormap.State{}}
	b.Field(name, o)
	b.values = append(b.values, o)
	return b
}

func (b *Builder) StartArrayField(name string) *Builder {
	a := Array{&list.State{}}
	b.Field(name, a)
	b.values = append(b.values, a)
	return b
}

func (b *Builder) Add(value Value) *Builder {
	if value == nil {
		value = Null{}
	}
	if b.root == nil {
		b.root = value
		return b
	}
	b.currentArray().v.Push(value, b.dot)
	return b
}

func (b *Builder) AddArray(values []Value) *Builder {
	var v Value
	if values == nil {
		v = Null{}
	} else {
		a := Array{&list.State{}}
		for _, v := range values {
			if v == nil {
				a.v.Push(Null{}, b.dot)
			} else {
				a.v.Push(v, b.dot)
			}
		}
		v = a
	}
	if b.root == nil {
		b.root = v
		return b
	}
	b.currentArray().v.Push(v, b.dot)
	return b
}

func (b *Builder) AddString(value string) *Builder {
	return b.Add(b.StringValue(value))
}

func (b *Builder) AddStringArray(values []string) *Builder {
	if values == nil {
		return b.Add(nil)
	}
	vs := make([]Value, len(values))
	for i, v := range values {
		vs[i] = b.StringValue(v)
	}
	return b.AddArray(vs)
}

func (b *Builder) AddBoolean(value bool) *Builder {
	return b.Add(Boolean(value))
}

func (b *Builder) AddBooleanArray(values []bool) *Builder {
	if values == nil {
		return b.Add(nil)
	}
	vs := make([]Value, len(values))
	for i, v := range values {
		vs[i] = Boolean(v)
	}
	return b.AddArray(vs)
}

func (b *Builder) AddInt(value int64) *Builder {
	return b.Add(Int(value))
}

func (b *Builder) AddIntArray(values []int64) *Builder {
	if values == nil {
		return b.Add(nil)
	}
	vs := make([]Value, len(values))
	for i, v := range values {
		vs[i] = Int(v)
	}
	return b.AddArray(vs)
}

func (b *Builder) AddFloat(value float64) *Builder {
	return b.Add(Float(value))
}

func (b *Builder) AddFloatArray(values []float64) *Builder {
	if values == nil {
		return b.Add(nil)
	}
	vs := make([]Value, len(values))
	for i, v := range values {
		vs[i] = Float(v)
	}
	return b.AddArray(vs)
}

func (b *Builder) currentObject() *Object {
	var o *Object
	if len(b.values) != 0 {
		v := b.values[len(b.values)-1]
		if x, ok := v.(Object); ok {
			o = &x
		}
	}
	if o == nil {
		panic("No current object")
	}
	return o
}

func (b *Builder) currentArray() *Array {
	var a *Array
	if len(b.values) != 0 {
		v := b.values[len(b.values)-1]
		if x, ok := v.(Array); ok {
			a = &x
		}
	}
	if a == nil {
		panic("No current array")
	}
	return a
}

func (b *Builder) build() (Value, error) {
	if b.root == nil {
		return nil, errors.New("Empty builder")
	}
	if len(b.values) != 0 {
		return nil, fmt.Errorf("%d objects or arrays not ended", len(b.values))
	}
	return b.root, nil
}

type Object struct {
	v *ormap.State
}

func (o Object) Value() interface{} {
	m := make(map[string]interface{})
	o.v.Entries(func(e ormap.Entry) {
		m[e.Key.(string)] = e.Value.(Value).Value()
	})
	return m
}

func (o Object) clone() Value {
	value := o.v.Transform(func(e *ormap.Element) *ormap.Element {
		return &ormap.Element{Value: e.Value.(Value).clone(), Dot: e.Dot}
	})
	return Object{value}
}

func (Object) isValue() {}

func (o Object) String() string {
	return fmt.Sprintf("%v", o.Value())
}

type Array struct {
	v *list.State
}

func (a Array) Value() interface{} {
	var vs []interface{}
	a.v.Values(func(x interface{}) {
		vs = append(vs, x.(Value).Value())
	})
	return vs
}

func (a Array) clone() Value {
	value := a.v.Transform(func(e *list.Element) *list.Element {
		return &list.Element{e.UID, e.Value.(Value).clone()}
	})
	return Array{value}
}

func (Array) isValue() {}

func (a Array) String() string {
	return fmt.Sprintf("%v", a.Value())
}

type String struct {
	v *text.State
}

func (s String) Value() interface{} {
	return s.v.String()
}

func (s String) clone() Value {
	return String{s.v.Clone()}
}

func (String) isValue() {}

func (s String) String() string {
	return fmt.Sprintf("%q", s.Value())
}

type Boolean bool

func (b Boolean) Value() interface{} {
	return bool(b)
}

func (b Boolean) clone() Value {
	return b
}

func (Boolean) isValue() {}

func (b Boolean) String() string {
	return fmt.Sprint(bool(b))
}

type Int int64

func (i Int) Value() interface{} {
	return int64(i)
}

func (i Int) clone() Value {
	return i
}

func (Int) isValue() {}

func (i Int) String() string {
	return fmt.Sprint(int64(i))
}

type Float float64

func (f Float) Value() interface{} {
	return float64(f)
}

func (f Float) clone() Value {
	return f
}

func (Float) isValue() {}

func (f Float) String() string {
	return fmt.Sprint(float64(f))
}

type Null struct{}

func (n Null) Value() interface{} {
	return nil
}

func (n Null) clone() Value {
	return n
}

func (Null) isValue() {}

func (Null) String() string {
	return "null"
}
