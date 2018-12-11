package crdt

import (
	"fmt"
	"reflect"
)

type JSONValue interface {
	Value() interface{}
	clone() JSONValue
	isJSONValue()
}

type JSONObject struct {
	v *mapState
}

func (v JSONObject) Value() interface{} {
	m := make(map[string]interface{})
	v.v.entries(func(e MapEntry) {
		m[e.Key.(string)] = e.Value.(JSONValue).Value()
	})
	return m
}

func (v JSONObject) clone() JSONValue {
	value := v.v.clone()
	for _, elements := range value.m {
		for i, e := range elements {
			elements[i] = &MapElement{e.Value.(JSONValue).clone(), e.Dot}
		}
	}
	return JSONObject{value}
}

func (JSONObject) isJSONValue() {}

type JSONArray struct {
	v *listState
}

func (v JSONArray) Value() interface{} {
	var vs []interface{}
	v.v.values(func(x interface{}) {
		vs = append(vs, x.(JSONValue).Value())
	})
	return vs
}

func (v JSONArray) clone() JSONValue {
	value := v.v.clone()
	for i, e := range value.elements {
		value.elements[i] = &ListElement{e.UID, e.Value.(JSONValue).clone()}
	}
	return JSONArray{value}
}

func (JSONArray) isJSONValue() {}

type JSONString struct {
	v *textState
}

func (v JSONString) Value() interface{} {
	return v.v.String()
}

func (v JSONString) clone() JSONValue {
	return JSONString{v.v.clone()}
}

func (JSONString) isJSONValue() {}

type JSONBoolean struct {
	V bool
}

func (v JSONBoolean) Value() interface{} {
	return v.V
}

func (v JSONBoolean) clone() JSONValue {
	return v
}

func (JSONBoolean) isJSONValue() {}

type JSONInt struct {
	V int64
}

func (v JSONInt) Value() interface{} {
	return v.V
}

func (v JSONInt) clone() JSONValue {
	return v
}

func (JSONInt) isJSONValue() {}

type JSONUint struct {
	V uint64
}

func (v JSONUint) Value() interface{} {
	return v.V
}

func (v JSONUint) clone() JSONValue {
	return v
}

func (JSONUint) isJSONValue() {}

type JSONFloat struct {
	V float64
}

func (v JSONFloat) Value() interface{} {
	return v.V
}

func (v JSONFloat) clone() JSONValue {
	return v
}

func (JSONFloat) isJSONValue() {}

type JSONNull struct {
}

func (v JSONNull) Value() interface{} {
	return nil
}

func (v JSONNull) clone() JSONValue {
	return v
}

func (JSONNull) isJSONValue() {}

var jsonNull = JSONNull{}

func toJSON(x reflect.Value, dot *Dot) (JSONValue, error) {
	if !x.IsValid() {
		return jsonNull, fmt.Errorf("Invalid JSON type: %s", x.Type())
	}
	if x.IsNil() {
		return jsonNull, nil
	}
	switch x.Kind() {
	case reflect.Bool:
		return JSONBoolean{x.Bool()}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return JSONInt{x.Int()}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return JSONUint{x.Uint()}, nil
	case reflect.Float32, reflect.Float64:
		return JSONFloat{x.Float()}, nil
	case reflect.String:
		value := &textState{tree: newTextTree()}
		s := x.String()
		if s != "" {
			value.replace(0, 0, s, dot)
		}
		return JSONString{value}, nil
	case reflect.Map:
		value := &mapState{}
		for _, k := range x.MapKeys() {
			if k.Kind() != reflect.String {
				return jsonNull, fmt.Errorf("Invalid JSON map key type: %s", k.Type())
			}
			v, err := toJSON(x.MapIndex(k), dot)
			if err != nil {
				return jsonNull, err
			}
			value.insert(k.String(), v, *dot)
		}
		return JSONObject{value}, nil
	case reflect.Array, reflect.Slice:
		value := &listState{}
		for i := 0; i < x.Len(); i++ {
			v, err := toJSON(x.Index(i), dot)
			if err != nil {
				return jsonNull, err
			}
			value.insert(i, v, dot)
		}
		return JSONArray{value}, nil
	default:
		return jsonNull, fmt.Errorf("Invalid JSON type: %s", x.Type())
	}
}

type JSONOp struct {
	Pointer []UID
	Inner   JSONInnerOp
}

func (j JSONOp) InsertedDots() []Dot {
	switch i := j.Inner.(type) {
	case JSONInnerObjectOp:
		if i.InsertedElement == nil {
			return nil
		}
		return []Dot{i.InsertedElement.Dot}
	case JSONInnerArrayOp:
		return i.Op.InsertedDots()
	case JSONInnerStringOp:
		var dots []Dot
		TextOp(i).InsertedDots(func(d *Dot) {
			dots = append(dots, *d)
		})
		return dots
	}
	panic("")
}

func (j JSONOp) Validate(siteID SiteID) bool {
	// TODO: Here
	return false
}

type LocalJSONOp interface {
	Pointer() []UID
	isLocalJSONOp()
}

type LocalJSONInsertOp struct {
	Value   interface{}
	pointer []UID
}

func (i LocalJSONInsertOp) Pointer() []UID {
	return i.pointer
}

func (LocalJSONInsertOp) isLocalJSONOp() {}

type LocalJSONRemoveOp struct {
	pointer []UID
}

func (r LocalJSONRemoveOp) Pointer() []UID {
	return r.pointer
}

func (LocalJSONRemoveOp) isLocalJSONOp() {}

type LocalJSONReplaceTextOp struct {
	Changes []TextEdit
	pointer []UID
}

func (t LocalJSONReplaceTextOp) Pointer() []UID {
	return t.pointer
}

func (LocalJSONReplaceTextOp) isLocalJSONOp() {}

type JSONUID interface {
	isJSONUID()
}

type JSONObjectUID struct {
	Key string
	Dot Dot
}

func (JSONObjectUID) isJSONUID() {}

type JSONArrayUID = UID

func (JSONArrayUID) isJSONUID() {}

type LocalJSONUID interface {
	isLocalJSONUID()
}

type LocalJSONObjectUID string

func (LocalJSONObjectUID) isLocalJSONUID() {}

type LocalJSONArrayUID int

func (LocalJSONArrayUID) isLocalJSONUID() {}

type JSONInnerOp interface {
	isJSONInnerOp()
}

type JSONInnerObjectOp MapOp

func (JSONInnerObjectOp) isJSONInnerOp() {}

type JSONInnerArrayOp struct {
	Op ListOp
}

func (JSONInnerArrayOp) isJSONInnerOp() {}

type JSONInnerStringOp TextOp

func (JSONInnerStringOp) isJSONInnerOp() {}
