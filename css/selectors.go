// Copyright 2014 Christer Sandberg.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.

package css

// CombinatorType identifies the combinator separating sequences of simple selectors.
// See http://www.w3.org/TR/selectors/#combinators
type Combinator int

const (
	Child Combinator = iota
	Descendant
	NextSibling
	LaterSibling
)

// Represents a pointer to the previous CompoundSelector separated by a Combinator.
type Prev struct {
	Combinator
	*CompoundSelector
}

// A CompoundSelector groups a SimpleSelector sequence and a pointer to the
// previous CompoundSelector separated by a Combinator if any.
// See http://www.w3.org/TR/selectors/#sequence
type CompoundSelector struct {
	SimpleSelectors []SimpleSelector
	*Prev
}

// Represents a selector.
// http://www.w3.org/TR/selectors/#selector-syntax
type Selector struct {
	CompoundSelector *CompoundSelector // A pointer to the last CompoundSelector.
	PseudoElement    *PseudoElementSelector
}

// SelectorsGroup represents a group of selectors.
// See http://www.w3.org/TR/selectors/#grouping
type SelectorsGroup []*Selector

// SimpleSelectorType identifies the type of simple selector.
type SimpleSelectorType int

// Type returns itself.
func (t SimpleSelectorType) Type() SimpleSelectorType {
	return t
}

const (
	Attribute SimpleSelectorType = iota
	LocalName
	PseudoClass
	PseudoElement
	PseudoNth
	PseudoNegation
	PseudoFunction
)

// Represents a simple selector.
// See http://www.w3.org/TR/selectors/#simple-selectors
type SimpleSelector interface {
	Type() SimpleSelectorType // The type of this simple selector.
}

// AttributeMatch identifies how to match an attribute.
type AttributeMatch int

const (
	Exists AttributeMatch = iota
	Equals
	Includes
	Begins
	Ends
	Contains
	Hyphens
)

// Represents an attribute selector.
// An attribute selector will also be used to represent class selectors and ID selectors.
// See http://www.w3.org/TR/selectors/#attribute-selectors
type AttributeSelector struct {
	SimpleSelectorType
	Match       AttributeMatch // How to match the attribute.
	Name, Value string         // Attribute name and value.
}

// Creates and returns a new AttributeSelector.
func NewAttributeSelector(match AttributeMatch, name, value string) *AttributeSelector {
	return &AttributeSelector{Attribute, match, name, value}
}

// Represents a type selector.
// See http://www.w3.org/TR/selectors/#type-selectors
type LocalNameSelector struct {
	SimpleSelectorType
	Name string // The tag name.
}

// Creates and returns a new LocalNameSelector.
func NewLocalNameSelector(name string) *LocalNameSelector {
	return &LocalNameSelector{LocalName, name}
}

// Represents a pseudo class selector.
// See http://www.w3.org/TR/selectors/#pseudo-classes
type PseudoClassSelector struct {
	SimpleSelectorType
	Value string // The value of this selector.
}

// Creates and returns a new PseudoClassSelector.
func NewPseudoClassSelector(value string) *PseudoClassSelector {
	return &PseudoClassSelector{PseudoClass, value}
}

// Represents a pseudo element selector.
// See http://www.w3.org/TR/selectors/#pseudo-elements
type PseudoElementSelector struct {
	SimpleSelectorType
	Value string // The value of this selector.
}

// Creates and returns a new PseudoElementSelector.
func NewPseudoElementSelector(value string) *PseudoElementSelector {
	return &PseudoElementSelector{PseudoElement, value}
}

// Represents a nth-* pseudo class selector.
// See http://www.w3.org/TR/selectors/#pseudo-classes
type PseudoNthSelector struct {
	SimpleSelectorType
	Name string // The name of this selector.
	A, B int    // The A and B arguments of this selector.
}

// Creates and returns a new PseudoNthSelector.
func NewPseudoNthSelector(name string, a, b int) *PseudoNthSelector {
	return &PseudoNthSelector{PseudoNth, name, a, b}
}

// Represents a functional pseudo class selector that is not
// of type :nth-* or :not.
// See http://www.w3.org/TR/selectors/#w3cselgrammar
type PseudoFunctionSelector struct {
	SimpleSelectorType
	Name, Arguments string // The name and arguments.
}

// Creates and returns a new PseudoFunctionSelector.
func NewPseudoFunctionSelector(name, arguments string) *PseudoFunctionSelector {
	return &PseudoFunctionSelector{PseudoFunction, name, arguments}
}

// Represents a negation pseudo class.
// See http://www.w3.org/TR/selectors/#negation
type PseudoNegationSelector struct {
	SimpleSelectorType
	Selector SimpleSelector
}

// Creates and returns a new PseudoNegationSelector
func NewPseudoNegationSelector(selector SimpleSelector) *PseudoNegationSelector {
	return &PseudoNegationSelector{PseudoNegation, selector}
}
