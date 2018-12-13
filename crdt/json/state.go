package json

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/chrsan/go/crdt"
	"github.com/chrsan/go/crdt/list"
	"github.com/chrsan/go/crdt/ormap"
	"github.com/chrsan/go/crdt/text"
)

type State struct {
	v Value
}

func (s *State) Insert(value interface{}, dot *crdt.Dot, pointer []interface{}) (Op, error) {
	if len(pointer) == 0 {
		return Op{nil, nil}, errors.New("Empty pointer")
	}
	p := pointer[len(pointer)-1]
	pointer = pointer[:len(pointer)-1]
	v, remotePointer, err := s.nestedLocal(pointer)
	if err != nil {
		return Op{nil, nil}, err
	}
	n, err := toValue(reflect.ValueOf(value), dot)
	if err != nil {
		return Op{nil, nil}, err
	}
	switch x := v.(type) {
	case Object:
		key, ok := p.(string)
		if !ok {
			return Op{nil, nil}, fmt.Errorf("Expected string key, got %T", p)
		}
		op := x.v.Insert(key, n, *dot)
		return Op{remotePointer, InnerObjectOp(op)}, nil
	case Array:
		i, ok := p.(int)
		if !ok {
			return Op{nil, nil}, fmt.Errorf("Expected int index, got %T", p)
		}
		if i < 0 || i >= x.v.Len() {
			return Op{nil, nil}, fmt.Errorf("Invalid array index: %d", i)
		}
		op := x.v.Insert(i, n, dot)
		return Op{remotePointer, InnerArrayOp{op}}, nil
	default:
		return Op{nil, nil}, fmt.Errorf("%v does not exist", pointer)
	}
}

func (s *State) Remove(pointer []interface{}) (Op, error) {
	if len(pointer) == 0 {
		return Op{nil, nil}, errors.New("Empty pointer")
	}
	p := pointer[0]
	pointer = pointer[1:]
	v, remotePointer, err := s.nestedLocal(pointer)
	if err != nil {
		return Op{nil, nil}, err
	}
	switch x := v.(type) {
	case Object:
		key, ok := p.(string)
		if !ok {
			return Op{nil, nil}, fmt.Errorf("Expected string key, got %T", p)
		}
		op, ok := x.v.Remove(key)
		if !ok {
			return Op{nil, nil}, fmt.Errorf("%v does not exist", pointer)
		}
		return Op{remotePointer, InnerObjectOp(op)}, nil
	case Array:
		i, ok := p.(int)
		if !ok {
			return Op{nil, nil}, fmt.Errorf("Expected int index, got %T", p)
		}
		if i < 0 || i >= x.v.Len() {
			return Op{nil, nil}, fmt.Errorf("Invalid array index: %d", i)
		}
		_, op := x.v.Remove(i)
		return Op{remotePointer, InnerArrayOp{op}}, nil
	default:
		return Op{nil, nil}, fmt.Errorf("%v does not exist", pointer)
	}
}

func (s *State) ReplaceText(index, count int, text string, dot *crdt.Dot, pointer []interface{}) (Op, error) {
	if len(pointer) == 0 {
		return Op{nil, nil}, errors.New("Empty pointer")
	}
	v, remotePointer, err := s.nestedLocal(pointer)
	if err != nil {
		return Op{nil, nil}, err
	}
	x, ok := v.(String)
	if !ok {
		return Op{nil, nil}, fmt.Errorf("Wrong type %T for pointer %#v", v, pointer)
	}
	op, ok := x.v.Replace(index, count, text, dot)
	if !ok {
		return Op{nil, nil}, fmt.Errorf("%v does not exist", pointer)
	}
	return Op{remotePointer, InnerStringOp(op)}, nil
}

func (s *State) ExecuteOp(op Op) (LocalOp, bool) {
	v, pointer, ok := s.nestedRemote(op.Pointer)
	if !ok {
		return nil, false
	}
	switch inner := op.Inner.(type) {
	case InnerObjectOp:
		x, ok := v.(Object)
		if !ok {
			return nil, false
		}
		e := x.v.Get(inner.Key)
		if e == nil && inner.InsertedElement == nil {
			return nil, false
		}
		lop := x.v.ExecuteOp(ormap.Op(inner))
		pointer = append(pointer, LocalObjectUID(lop.Key.(string)))
		if lop.IsInsert {
			return LocalInsertOp{lop.Value.(Value).Value(), pointer}, true
		}
		return LocalRemoveOp(pointer), true
	case InnerArrayOp:
		x, ok := v.(Array)
		if !ok {
			return nil, false
		}
		lop, ok := x.v.ExecuteOp(inner.Op)
		if !ok {
			return nil, false
		}
		switch lop := lop.(type) {
		case list.LocalInsertOp:
			pointer = append(pointer, LocalArrayUID(lop.Index))
			return LocalInsertOp{lop.Value.(Value).Value(), pointer}, true
		case list.LocalRemoveOp:
			pointer = append(pointer, LocalArrayUID(lop))
			return LocalRemoveOp(pointer), true
		default:
			panic(lop)
		}
	case InnerStringOp:
		x, ok := v.(String)
		if !ok {
			return nil, false
		}
		changes := x.v.ExecuteOp(text.Op(inner))
		if len(changes) == 0 {
			return nil, false
		}
		return LocalReplaceTextOp{changes, pointer}, true
	default:
		panic(inner)
	}
}

func (s *State) NestedValue(pointer []interface{}) (Value, error) {
	if len(pointer) == 0 {
		return nil, errors.New("Empty pointer")
	}
	v := s.v
	for _, p := range pointer {
		switch x := v.(type) {
		case Object:
			key, ok := p.(string)
			if !ok {
				return nil, fmt.Errorf("Expected string key, got %T", p)
			}
			e := x.v.Get(key)
			if e == nil {
				return nil, nil
			}
			v = e.Value.(Value)
		case Array:
			i, ok := p.(int)
			if !ok {
				return nil, fmt.Errorf("Expected int index, got %T", p)
			}
			if i < 0 || i >= x.v.Len() {
				return nil, fmt.Errorf("Invalid array index: %d", i)
			}
			e := x.v.Get(i)
			if e == nil {
				return nil, nil
			}
			v = e.Value.(Value)
		default:
			return nil, nil
		}
	}
	return v, nil
}

func (s *State) Clone() *State {
	return &State{s.v.clone()}
}

func (s *State) Eq(state *State) bool {
	switch x := s.v.(type) {
	case Object:
		if y, ok := state.v.(Object); ok {
			return x.v.Eq(y.v)
		}
		return false
	case Array:
		if y, ok := state.v.(Array); ok {
			return x.v.Eq(y.v)
		}
		return false
	case String:
		if y, ok := state.v.(String); ok {
			return x.v.Eq(y.v)
		}
		return false
	case Boolean:
		if y, ok := state.v.(Boolean); ok {
			return x == y
		}
		return false
	case Int:
		if y, ok := state.v.(Int); ok {
			return x == y
		}
		return false
	case Uint:
		if y, ok := state.v.(Uint); ok {
			return x == y
		}
		return false
	case Float:
		if y, ok := state.v.(Float); ok {
			return x == y
		}
		return false
	case Null:
		if _, ok := state.v.(Null); ok {
			return true
		}
		return false
	default:
		panic(x)
	}
}

func (s *State) nestedLocal(pointer []interface{}) (Value, []UID, error) {
	v := s.v
	var remotePointer []UID
	for _, p := range pointer {
		switch x := v.(type) {
		case Object:
			key, ok := p.(string)
			if !ok {
				return nil, nil, fmt.Errorf("Expected string key, got %T", p)
			}
			e := x.v.Get(key)
			if e == nil {
				return nil, nil, fmt.Errorf("%v does not exist", pointer)
			}
			uid := ObjectUID{key, e.Dot}
			remotePointer = append(remotePointer, uid)
			v = e.Value.(Value)
		case Array:
			i, ok := p.(int)
			if !ok {
				return nil, nil, fmt.Errorf("Expected int index, got %T", p)
			}
			if i < 0 || i >= x.v.Len() {
				return nil, nil, fmt.Errorf("Invalid array index: %d", i)
			}
			e := x.v.Get(i)
			if e == nil {
				return nil, nil, fmt.Errorf("%v does not exist", pointer)
			}
			uid := ArrayUID(e.UID)
			remotePointer = append(remotePointer, uid)
			v = e.Value.(Value)
		default:
			return nil, nil, fmt.Errorf("%v does not exist", pointer)
		}
	}

	return v, remotePointer, nil
}

func (s *State) nestedRemote(pointer []UID) (Value, []LocalUID, bool) {
	v := s.v
	var localPointer []LocalUID
	for _, u := range pointer {
		switch x := v.(type) {
		case Object:
			u, ok := u.(ObjectUID)
			if !ok {
				return nil, nil, false
			}
			e := x.v.GetElement(u.Key, &u.Dot)
			if e == nil {
				return nil, nil, false
			}
			localPointer = append(localPointer, LocalObjectUID(u.Key))
			v = e.Value.(Value)
		case Array:
			u, ok := u.(ArrayUID)
			if !ok {
				return nil, nil, false
			}
			uid := crdt.UID(u)
			i, ok := x.v.Index(&uid)
			if !ok {
				return nil, nil, false
			}
			localPointer = append(localPointer, LocalArrayUID(i))
			v = x.v.Get(i).Value.(Value)
		default:
			return nil, nil, false
		}
	}

	return v, localPointer, true
}
