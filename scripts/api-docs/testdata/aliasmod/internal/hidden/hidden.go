package hidden

import (
	"context"
	left "example.com/aliasmod/examples/left"
	right "example.com/aliasmod/examples/right"
)

type State int

const (
	Ready State = iota
	Done
)

const (
	PriceLabel   = "costs $5"
	InternalName = "renamed costs $5"
	Count        = 5
	Enabled      = true
)

type Base interface {
	// Ping verifies that the component is responsive.
	Ping() error
}

type Composite interface {
	Base
}

type Embedded struct {
	Enabled bool `json:"enabled"`
}

type Item struct {
	// Name is the externally visible name and costs $5.
	Name string `json:"name"`
	// State is the current lifecycle state.
	State State `json:"state"`
	// Embedded contributes its exported fields.
	Embedded
	Left   left.Token
	Right  right.Token
	secret string
}

// Rename changes the item's public name.
func (i *Item) Rename(ctx context.Context, name string) error { return nil }
