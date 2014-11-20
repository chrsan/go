// Copyright 2014 Christer Sandberg.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.

package css

import "golang.org/x/net/html"

// Returns all the nodes within n that match the given selectors string.
func QuerySelectorAll(selectors string, n *html.Node) ([]*html.Node, error) {
	s, err := ParseSelectorFromString(selectors)
	if err != nil {
		return nil, err
	}

	return QueryAll(s, n), nil
}

// Returns all the nodes within n that match the given SelectorsGroup.
func QueryAll(s SelectorsGroup, n *html.Node) []*html.Node {
	var result []*html.Node
	Traverse(n, func(x *html.Node) {
		if MatchesSelectors(s, x, nil) {
			result = append(result, x)
		}
	})

	return result
}

// Traverses the nodes within n using depth-first pre-order traversal.
func Traverse(n *html.Node, f func(*html.Node)) {
	if n.Type == html.ElementNode {
		f(n)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		Traverse(c, f)
	}
}
