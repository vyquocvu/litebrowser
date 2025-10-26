package css

// StyleSheet represents a CSS stylesheet.
type StyleSheet struct {
	Rules []Rule
}

// Rule represents a single CSS rule.
type Rule struct {
	Selectors []Selector
	Declarations []Declaration
}

// Selector represents a CSS selector.
// For now, we'll support simple selectors.
type Selector struct {
	TagName string
	ID      string
	Classes []string
}

// Declaration represents a CSS property-value pair.
type Declaration struct {
	Property string
	Value    string
}
