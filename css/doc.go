// Copyright 2014 Christer Sandberg.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.

/*
Package css implements a CSS tokenizer and selector parser according to the Selector Level 3 specificaton,
and it provides functions to match selectors against HTML nodes.

Use the high-level API to match selectors the same way you would in JavaScript:

	doc := html.Parse(...)
	nodes, err := QuerySelectorAll(`div:first-child`, doc)

The low-level API can be used to gain more control of the matching process.
Here's an example of permitting the deprecated :contains pseudo class:

	doc := html.Parse(...)
	// Create a SimpleSelectorMatchFunc
	m := func(ss SimpleSelector, n *html.Node) bool {
		s, ok := ss.(*PseudoFunctionSelector)
		if !ok || s.Name != "contains" {
			return false
		}

		return strings.Contains(text(n), s.Arguments)
	}

	s, _ := ParseSelectorFromString(`div:contains('foo')`)

	var r []*html.Node
	Traverse(doc, func(n *html.Node) {
		if MatchesSelectors(s, n, m) {
			r = append(r, n)
		}
	})

The Tokenizer is a full blown CSS tokenizer and isn't limited to tokenizing what's specified in the Selector specification.
*/
package css
