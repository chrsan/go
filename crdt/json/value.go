package json

import (
	"fmt"
	"reflect"

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

// TODO: Check nil

func toValue(x reflect.Value, dot *crdt.Dot) (Value, error) {
	if !x.IsValid() {
		return null, fmt.Errorf("Invalid JSON type: %s", x.Type())
	}
	switch x.Kind() {
	case reflect.Bool:
		return Boolean(x.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return Int(x.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return Uint(x.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return Float(x.Float()), nil
	case reflect.String:
		value := text.NewState()
		s := x.String()
		if s != "" {
			value.Replace(0, 0, s, dot)
		}
		return String{value}, nil
	case reflect.Map:
		value := &ormap.State{}
		for _, k := range x.MapKeys() {
			if k.Kind() != reflect.String {
				return null, fmt.Errorf("Invalid JSON map key type: %s", k.Type())
			}
			v, err := toValue(x.MapIndex(k), dot)
			if err != nil {
				return null, err
			}
			value.Insert(k.String(), v, *dot)
		}
		return Object{value}, nil
	case reflect.Array, reflect.Slice:
		value := &list.State{}
		for i := 0; i < x.Len(); i++ {
			v, err := toValue(x.Index(i), dot)
			if err != nil {
				return null, err
			}
			value.Insert(i, v, dot)
		}
		return Array{value}, nil
	default:
		return null, fmt.Errorf("Invalid JSON type: %s", x.Type())
	}
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

type Uint uint64

func (u Uint) Value() interface{} {
	return uint64(u)
}

func (u Uint) clone() Value {
	return u
}

func (Uint) isValue() {}

func (u Uint) String() string {
	return fmt.Sprint(uint64(u))
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

var null = Null{}
