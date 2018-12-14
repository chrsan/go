package json

import (
	"encoding/json"
	"io"
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

func FromBuilder(siteID crdt.SiteID, f func(*Builder)) (*JSON, error) {
	crdt.CheckSiteID(siteID)
	summary := crdt.NewSummary()
	dot := summary.Dot(siteID)
	b := Builder{dot: &dot}
	f(&b)
	v, err := b.build()
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
	var keys []string
	var values []Value
	push := func(v Value) {
		x := values[len(values)-1]
		if arr, ok := x.(Array); ok {
			arr.v.Push(v, &dot)
		} else {
			inKey = true
			obj := x.(Object)
			key := keys[len(keys)-1]
			keys = keys[:len(keys)-1]
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
				values = append(values, Object{&ormap.State{}})
			case '}':
				if len(values) > 1 {
					v := values[len(values)-1]
					values = values[:len(values)-1]
					push(v)
				}
			case '[':
				values = append(values, Array{&list.State{}})
			case ']':
				if len(values) > 1 {
					v := values[len(values)-1]
					values = values[:len(values)-1]
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
			if len(values) == 0 {
				return &JSON{siteID, &State{v}, summary}, nil
			}
			push(v)
		case string:
			if inKey {
				inKey = false
				keys = append(keys, x)
				continue
			}
			t := text.NewState()
			if x != "" {
				t.Replace(0, 0, x, &dot)
			}
			if len(values) == 0 {
				return &JSON{siteID, &State{String{t}}, summary}, nil
			}
			push(String{t})
		case bool:
			if len(values) == 0 {
				return &JSON{siteID, &State{Boolean(x)}, summary}, nil
			}
			push(Boolean(x))
		case nil:
			if len(values) == 0 {
				return &JSON{siteID, &State{Null{}}, summary}, nil
			}
			push(Null{})
		}
	}
	if len(values) == 0 {
		return nil, nil
	}
	return &JSON{siteID, &State{values[0]}, summary}, nil
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

func (j *JSON) Insert(value Value, pointer ...interface{}) (Op, error) {
	dot := j.summary.Dot(j.siteID)
	return j.state.Insert(value, &dot, pointer)
}

func (j *JSON) InsertBuilder(f func(*Builder), pointer ...interface{}) (Op, error) {
	dot := j.summary.Dot(j.siteID)
	b := Builder{dot: &dot}
	f(&b)
	v, err := b.build()
	if err != nil {
		return Op{nil, nil}, err
	}
	return j.state.Insert(v, &dot, pointer)
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
	if v == nil {
		return nil, nil
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
