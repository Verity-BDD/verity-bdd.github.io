package facade

import impl "example.com/aliasmod/internal/hidden"

// Item is public.
type Item = impl.Item

// State is public.
type State = impl.State

// Embedded is public.
type Embedded = impl.Embedded

// Composite exposes inherited methods.
type Composite = impl.Composite

const (
	Ready         = impl.Ready
	Done          = impl.Done
	GroupedLabel  = impl.PriceLabel
	GroupedCount  = impl.Count
	GroupedSwitch = impl.Enabled
)

// PriceLabel is a standalone forwarded constant whose value contains $5.
const PriceLabel = impl.PriceLabel

// TypedPriceLabel is an explicitly typed standalone forwarded constant.
const TypedPriceLabel string = impl.PriceLabel

// TypedCount is an explicitly typed numeric forwarded constant.
const TypedCount int = impl.Count

// TypedSwitch is an explicitly typed boolean forwarded constant.
const TypedSwitch bool = impl.Enabled

// RenamedPriceLabel forwards a differently named implementation constant.
const RenamedPriceLabel = impl.InternalName

// LiteralPriceLabel is already public and must remain untouched.
const LiteralPriceLabel = "literal costs $5"

var Default = impl.Item{Name: "default"}
