// Copyright 2014 Christer Sandberg.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.

package css

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"strconv"
	"testing"
)

func TestTokenizerSpecCompliance(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/component_value_list.json")
	if err != nil {
		t.Fatal("Could not read component_value_list.json")
	}

	var arr []interface{}
	if err := json.Unmarshal(data, &arr); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < len(arr); i += 2 {
		input := arr[i].(string)
		r := tokenize(input, t)
		e := arr[i+1].([]interface{})
		if len(r) != len(e) {
			t.Errorf(`Expected len %v for test %v, got %v`, len(e), len(r), i)
			continue
		}

		for j, v := range e {
			if !reflect.DeepEqual(r[j], v) {
				t.Logf("Value mismatch at array position %v for test %v:\n%q", j, i, input)
				t.Errorf("%#v != %#v", r[j], v)
			}
		}
	}
}

func tokenize(input string, t *testing.T) []interface{} {
	var r []interface{}

	tokenizer := NewTokenizer(input)
	for {
		var v interface{}
		tk := tokenizer.NextToken()
		switch tk.Type() {
		case EOF:
			return r
		case AtKeyword:
			v = vals("at-keyword", tk.String())
		case BadString:
			v = vals("error", "bad-string")
		case BadUrl:
			v = vals("error", "bad-url")
		case Function:
			v = vals("function", tk.String())
		case Hash:
			h := tk.(*HashToken)
			s := "unrestricted"
			if h.ID {
				s = "id"
			}

			v = vals("hash", h.Value, s)
		case Ident:
			v = vals("ident", tk.String())
		case String:
			v = vals("string", tk.String())
		case URL:
			v = vals("url", tk.String())
		case Whitespace:
			v = " "
		case Dimension:
			d := tk.(*DimensionToken)
			s := "number"
			if d.Integer {
				s = "integer"
			}

			f, err := strconv.ParseFloat(d.Value, 64)
			if err != nil {
				t.Fatal(err)
			}

			v = vals("dimension", d.Value, f, s, d.Unit)
		case Number, Percentage:
			n := tk.(*NumberToken)
			s := "number"
			if n.Integer {
				s = "integer"
			}

			f, err := strconv.ParseFloat(n.Value, 64)
			if err != nil {
				t.Fatal(err)
			}

			x := "number"
			if n.TokenType == Percentage {
				x = "percentage"
			}

			v = vals(x, n.Value, f, s)
		case UnicodeRange:
			u := tk.(*UnicodeRangeToken)
			v = vals("unicode-range", float64(u.Start), float64(u.End))
		default:
			v = tk.String()
		}

		r = append(r, v)
	}
}

func vals(v ...interface{}) []interface{} {
	return v
}
