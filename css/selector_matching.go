// Copyright 2014 Christer Sandberg.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.

package css

import (
	"strings"

	"golang.org/x/net/html"
)

// Represents a callback function that will be invoked when the default
// matching machinery couldn't match the given SimpleSelector and HTML node.
// If the callback function returns true it means that the node matches.
type SimpleSelectorMatchFunc func(SimpleSelector, *html.Node) bool

// Matches the given SelectorGroup s against the HTML node n using the callback function
// f if the default matching machinery didn't find a match.
func MatchesSelectors(s SelectorsGroup, n *html.Node, f SimpleSelectorMatchFunc) bool {
	for _, x := range s {
		if MatchesSelector(x, n, f) {
			return true
		}
	}

	return false
}

// Matches the given Selector s against the HTML node n using the callback function
// f if the default matching machinery didn't find a match.
func MatchesSelector(s *Selector, n *html.Node, f SimpleSelectorMatchFunc) bool {
	return s.PseudoElement == nil && matchesCompoundSelector(s.CompoundSelector, n, f) == matched
}

// Matches the given SimpleSelector s against the HTML node n using the callback function
// f if the default matching machinery didn't find a match.
func MatchesSimpleSelector(s SimpleSelector, n *html.Node, f SimpleSelectorMatchFunc) bool {
	if n.Type == html.DocumentNode {
		for n = n.FirstChild; n != nil; n = n.NextSibling {
			if n.Type == html.ElementNode {
				break
			}
		}
	}

	if n.Type != html.ElementNode {
		return false
	}

	switch x := s.(type) {
	case *LocalNameSelector:
		return strings.EqualFold(n.Data, x.Name)
	case *AttributeSelector:
		return matchesAttributeSelector(x, n)
	case *PseudoNegationSelector:
		return !MatchesSimpleSelector(x.Selector, n, f)
	case *PseudoClassSelector:
		if matchesPseudoClassSelector(x, n) {
			return true
		}
	case *PseudoNthSelector:
		if matchesPseudoNthSelector(x, n) {
			return true
		}
	}

	if f != nil {
		return f(s, n)
	}

	return false
}

type matchingResult int

const (
	matched matchingResult = iota
	notMatched
	restartFromClosestDescendant
	restartFromClosestLaterSibling
)

func matchesCompoundSelector(s *CompoundSelector, n *html.Node, f SimpleSelectorMatchFunc) matchingResult {
	for _, ss := range s.SimpleSelectors {
		if !MatchesSimpleSelector(ss, n, f) {
			return restartFromClosestLaterSibling
		}
	}

	if s.Prev == nil {
		return matched
	}

	siblings := false
	candidateNotFound := notMatched

	switch s.Prev.Combinator {
	case NextSibling, LaterSibling:
		siblings = true
		candidateNotFound = restartFromClosestDescendant
	}

	for {
		var nextNode *html.Node
		if siblings {
			nextNode = n.PrevSibling
		} else {
			nextNode = n.Parent
		}

		if nextNode == nil {
			return candidateNotFound
		} else {
			n = nextNode
		}

		if n.Type == html.ElementNode {
			r := matchesCompoundSelector(s.Prev.CompoundSelector, n, f)
			if r == matched || r == notMatched {
				return r
			}

			switch s.Prev.Combinator {
			case Child:
				return restartFromClosestDescendant
			case NextSibling:
				return r
			case LaterSibling:
				if r == restartFromClosestDescendant {
					return r
				}
			}
		}
	}
}

func matchesAttributeSelector(s *AttributeSelector, n *html.Node) bool {
	for _, a := range n.Attr {
		if a.Key != s.Name {
			continue
		}

		switch s.Match {
		case Exists:
			return true
		case Equals:
			return a.Val == s.Value
		case Includes:
			for _, x := range strings.FieldsFunc(a.Val, IsSpace) {
				if x == s.Value {
					return true
				}
			}
		case Begins:
			return strings.HasPrefix(a.Val, s.Value)
		case Ends:
			return strings.HasSuffix(a.Val, s.Value)
		case Contains:
			return strings.Contains(a.Val, s.Value)
		case Hyphens:
			return a.Val == s.Value || strings.HasPrefix(a.Val, s.Value+"-")
		}

		return false
	}

	return false
}

func matchesPseudoClassSelector(s *PseudoClassSelector, n *html.Node) bool {
	switch s.Value {
	case "first-child":
		return matchesFirstOrLastChild(n, true)
	case "last-child":
		return matchesFirstOrLastChild(n, false)
	case "only-child":
		return matchesFirstOrLastChild(n, true) && matchesFirstOrLastChild(n, false)
	case "first-of-type":
		return matchesNthChild(n, 0, 1, true, false)
	case "last-of-type":
		return matchesNthChild(n, 0, 1, true, true)
	case "only-of-type":
		return matchesNthChild(n, 0, 1, true, false) && matchesNthChild(n, 0, 1, true, true)
	case "root":
		return n.Parent != nil && n.Parent.Type == html.DocumentNode
	case "empty":
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				return false
			}

			if c.Type == html.TextNode && len(c.Data) > 0 {
				return false
			}
		}

		return true
	default:
		return false
	}
}

func matchesPseudoNthSelector(s *PseudoNthSelector, n *html.Node) bool {
	switch s.Name {
	case "nth-child":
		return matchesNthChild(n, s.A, s.B, false, false)
	case "nth-last-child":
		return matchesNthChild(n, s.A, s.B, false, true)
	case "nth-of-type":
		return matchesNthChild(n, s.A, s.B, true, false)
	case "nth-last-of-type":
		return matchesNthChild(n, s.A, s.B, true, true)
	default:
		return false
	}
}

func matchesFirstOrLastChild(n *html.Node, first bool) bool {
	for {
		var x *html.Node
		if first {
			x = n.PrevSibling
		} else {
			x = n.NextSibling
		}

		if x == nil {
			x = n.Parent
			return x != nil && x.Type != html.DocumentNode
		} else {
			if x.Type == html.ElementNode {
				return false
			}

			n = x
		}
	}
}

func matchesNthChild(n *html.Node, a, b int, isOfType, fromEnd bool) bool {
	if n.Parent == nil || n.Parent.Type == html.DocumentNode {
		return false
	}

	x := n
	i := 1
	for {
		if fromEnd {
			s := x.NextSibling
			if s == nil {
				break
			} else {
				x = s
			}
		} else {
			s := x.PrevSibling
			if s == nil {
				break
			} else {
				x = s
			}
		}

		if x.Type == html.ElementNode {
			if isOfType {
				if n.Data == x.Data {
					i += 1
				}
			} else {
				i += 1
			}
		}
	}

	if a == 0 {
		return b == i
	}

	return ((i-b)/a) >= 0 && ((i-b)%a) == 0
}
