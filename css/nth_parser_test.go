// Copyright 2014 Christer Sandberg.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.

package css

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestNthParserSpecCompliance(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/An+B.json")
	if err != nil {
		t.Fatal("Could not read An+B.json")
	}

	var arr []interface{}
	if err := json.Unmarshal(data, &arr); err != nil {
		t.Fatal(err)
	}

	if len(arr)%2 != 0 {
		t.Fatal("An+B.json is invalid")
	}

	var args string
	var tok Tokenizer
	for i, v := range arr {
		if i%2 == 0 {
			args = v.(string)
			tok = NewTokenizer(args + ")")
		} else {
			if v == nil {
				_, _, err := parseNth(tok)
				if err == nil {
					t.Errorf(`Expected error for nth arguments %q at index %v`, args, i)
				}
			} else {
				n := v.([]interface{})
				a1, b1 := n[0].(float64), n[1].(float64)
				a2, b2, err := parseNth(tok)
				if err != nil {
					t.Errorf(`Error parsing nth arguments %q at index %v`, args, i)
					continue
				}

				if float64(a2) != a1 || float64(b2) != b1 {
					t.Errorf(`Value mismatch at index %v: [%v %v] != [%v %v]`, i, a2, b2, a1, b1)
				}
			}
		}
	}
}
