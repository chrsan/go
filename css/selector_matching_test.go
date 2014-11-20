// Copyright 2014 Christer Sandberg.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.

package css

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

var testSelectors = map[string]int{
	`*`:                           251,
	`:root`:                       1,
	`:empty`:                      2,
	`div:first-child`:             51,
	`div:nth-child(even)`:         106,
	`div:nth-child(2n)`:           106,
	`div:nth-child(odd)`:          137,
	`div:nth-child(2n+1)`:         137,
	`div:nth-child(n)`:            243,
	`script:first-of-type`:        1,
	`div:last-child`:              53,
	`script:last-of-type`:         1,
	`script:nth-last-child(odd)`:  1,
	`script:nth-last-child(even)`: 1,
	`script:nth-last-child(5)`:    0,
	`script:nth-of-type(2)`:       1,
	`script:nth-last-of-type(n)`:  2,
	`div:only-child`:              22,
	`meta:only-of-type`:           1,
	`div > div`:                   242,
	`div + div`:                   190,
	`div ~ div`:                   190,
	`body`:                        1,
	`body div`:                    243,
	`div`:                         243,
	`div div`:                     242,
	`div div div`:                 241,
	`div, div, div`:               243,
	`div, a, span`:                243,
	`.dialog`:                     51,
	`div.dialog`:                  51,
	`div .dialog`:                 51,
	`div.character, div.dialog`:   99,
	`#speech5`:                    1,
	`div#speech5`:                 1,
	`div #speech5`:                1,
	`div.scene div.dialog`:        49,
	`div#scene1 div.dialog div`:   142,
	`#scene1 #speech1`:            1,
	`div[class]`:                  103,
	`div[class=dialog]`:           50,
	`div[class^=dia]`:             51,
	`div[class$=log]`:             50,
	`div[class*=sce]`:             1,
	`div[class|=dialog]`:          50,
	`div[class~=dialog]`:          51,
	`head > :not(meta)`:           2,
	`head > :not(:last-child)`:    2,
}

var dom *html.Node

func init() {
	f, err := os.Open("testdata/test.html")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	dom, err = html.Parse(f)
	if err != nil {
		panic(err)
	}
}

func TestSelectorMatching(t *testing.T) {
	for k, v := range testSelectors {
		if r, err := QuerySelectorAll(k, dom); err != nil {
			t.Errorf(`Could not parse selector %q (%s)`, k, err)
		} else if len(r) != v {
			t.Errorf(`Got %v nodes matching %q, want %v`, len(r), k, v)
		}
	}
}

func TestSelectorMatchingWithMatchFunc(t *testing.T) {
	m := func(s SelectorsGroup) int {
		c := 0
		Traverse(dom, func(n *html.Node) {
			if MatchesSelectors(s, n, containsMatcher) {
				c++
			}
		})

		return c
	}

	s1, _ := ParseSelectorFromString(`h3:contains('palace')`)
	c1 := m(s1)
	if c1 != 1 {
		t.Errorf(`Got %v nodes matching "h3:contains('palace'), want 1`, c1)
	}

	s2, _ := ParseSelectorFromString(`:contains('Boom')`)
	c2 := m(s2)
	if c2 != 0 {
		t.Errorf(`Got %v nodes matching "div:contains('Boom'), want 0`, c2)
	}

}

func containsMatcher(s SimpleSelector, n *html.Node) bool {
	fs, ok := s.(*PseudoFunctionSelector)
	if !ok || fs.Name != "contains" {
		return false
	}

	var b bytes.Buffer
	var f func(*html.Node, *bytes.Buffer)
	f = func(n *html.Node, b *bytes.Buffer) {
		switch n.Type {
		case html.TextNode:
			b.WriteString(n.Data)
		case html.ElementNode:
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c, b)
			}
		}
	}

	f(n, &b)
	return strings.Contains(b.String(), fs.Arguments)
}
