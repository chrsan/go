package json

import (
	"encoding/json"
	"io"
	"reflect"
	"strings"

	"github.com/chrsan/go/crdt/ormap"

	"github.com/chrsan/go/crdt"
	"github.com/chrsan/go/crdt/list"
	"github.com/chrsan/go/crdt/text"
)

type JSON struct {
	siteID  crdt.SiteID
	state   *State
	summary *crdt.Summary
}

func New(siteID crdt.SiteID, value interface{}) (*JSON, error) {
	crdt.CheckSiteID(siteID)
	summary := crdt.NewSummary()
	dot := summary.Dot(siteID)
	if value == nil {
		return &JSON{siteID, &State{null}, summary}, nil
	}
	v, err := toValue(reflect.ValueOf(value), &dot)
	if err != nil {
		return nil, err
	}
	return &JSON{siteID, &State{v}, summary}, nil
}

func FromString(siteID crdt.SiteID, s string) (*JSON, error) {
	crdt.CheckSiteID(siteID)
	summary := crdt.NewSummary()
	dot := summary.Dot(siteID)
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	inKey := false
	var keyStack []string
	var valueStack []Value
	push := func(v Value) {
		x := valueStack[len(valueStack)-1]
		if arr, ok := x.(Array); ok {
			arr.v.Push(v, &dot)
		} else {
			inKey = true
			obj := x.(Object)
			key := keyStack[len(keyStack)-1]
			keyStack = keyStack[:len(keyStack)-1]
			obj.v.Insert(key, v, dot)
		}
	}
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch x := tok.(type) {
		case json.Delim:
			switch x {
			case '{':
				inKey = true
				valueStack = append(valueStack, Object{&ormap.State{}})
			case '}':
				if len(valueStack) != 1 {
					v := valueStack[len(valueStack)-1]
					valueStack = valueStack[:len(valueStack)-1]
					push(v)
				}
			case '[':
				valueStack = append(valueStack, Array{&list.State{}})
			case ']':
				if len(valueStack) != 1 {
					v := valueStack[len(valueStack)-1]
					valueStack = valueStack[:len(valueStack)-1]
					push(v)
				}
			}
		case json.Number:
			var v Value
			if strings.LastIndexByte(x.String(), '.') != -1 {
				f, _ := x.Float64()
				v = Float(f)
			} else {
				i, _ := x.Int64()
				v = Int(i)
			}
			if len(valueStack) == 0 {
				return &JSON{siteID, &State{v}, summary}, nil
			}
			push(v)
		case string:
			if inKey {
				inKey = false
				keyStack = append(keyStack, x)
				continue
			}
			t := text.NewState()
			if x != "" {
				t.Replace(0, 0, x, &dot)
			}
			if len(valueStack) == 0 {
				return &JSON{siteID, &State{String{t}}, summary}, nil
			}
			push(String{t})
		case bool:
			if len(valueStack) == 0 {
				return &JSON{siteID, &State{Boolean(x)}, summary}, nil
			}
			push(Boolean(x))
		case nil:
			if len(valueStack) == 0 {
				return &JSON{siteID, &State{null}, summary}, nil
			}
			push(null)
		}
	}
	if len(valueStack) == 0 {
		return nil, nil
	}
	return &JSON{siteID, &State{valueStack[0]}, summary}, nil
}

func (j *JSON) Len(pointer ...interface{}) (int, error) {
	v, err := j.state.NestedValue(pointer)
	if err != nil {
		return 0, err
	}
	switch x := v.(type) {
	case Object:
		return x.v.Len(), nil
	case Array:
		return x.v.Len(), nil
	case String:
		return x.v.Len(), nil
	default:
		return 0, nil
	}
}

func (j *JSON) Insert(value interface{}, pointer ...interface{}) (Op, error) {
	dot := j.summary.Dot(j.siteID)
	return j.state.Insert(value, &dot, pointer)
}

func (j *JSON) Remove(pointer ...interface{}) (Op, error) {
	return j.state.Remove(pointer)
}

func (j *JSON) ReplaceText(index, count int, text string, pointer ...interface{}) (Op, error) {
	dot := j.summary.Dot(j.siteID)
	return j.state.ReplaceText(index, count, text, &dot, pointer)
}

func (j *JSON) ExecuteOp(op Op) (LocalOp, bool) {
	switch inner := op.Inner.(type) {
	case InnerObjectOp:
		if inner.InsertedElement != nil {
			j.summary.Insert(&inner.InsertedElement.Dot)
		}
	case InnerArrayOp:
		e, ok := inner.Op.InsertedElement()
		if ok {
			j.summary.Insert(&e.UID.Dot)
		}
	case InnerStringOp:
		for i := range inner.RemovedUIDs {
			uid := &inner.RemovedUIDs[i]
			j.summary.Insert(&uid.Dot)
		}
	default:
		panic(op)
	}
	return j.state.ExecuteOp(op)
}

func (j *JSON) Value(pointer ...interface{}) (interface{}, error) {
	if len(pointer) == 0 {
		return j.state.v.Value(), nil
	}
	v, err := j.state.NestedValue(pointer)
	if err != nil {
		return nil, err
	}
	return v.Value(), nil
}

func (j *JSON) Replicate(siteID crdt.SiteID) *JSON {
	crdt.CheckSiteID(siteID)
	return &JSON{siteID, j.state.Clone(), j.summary.Clone()}
}

func (j *JSON) Eq(json *JSON) bool {
	return j.state.Eq(json.state) && j.summary.Eq(json.summary)
}
