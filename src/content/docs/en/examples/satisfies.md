---
title: Custom Expectations (Satisfies)
description: Examples of writing custom expectations with the Satisfies pattern
---

The `Satisfies` expectation allows you to create custom validation logic using functions. It's particularly useful for complex business rules, struct comparisons, and scenarios where built-in expectations aren't sufficient.

## Basic Usage

### Simple Value Validation

```go
actor.AttemptsTo(
    ensure.That(answerable.ValueOf(age), expectations.Satisfies("is positive number", func(actual int) error {
        if actual <= 0 {
            return fmt.Errorf("expected positive value, but got %d", actual)
        }
        return nil
    })),
)

actor.AttemptsTo(
    ensure.That(answerable.ValueOf(email), expectations.Satisfies("contains valid email", func(actual string) error {
        if !strings.Contains(actual, "@") {
            return fmt.Errorf("missing @ symbol in email")
        }
        if !strings.Contains(actual, ".") {
            return fmt.Errorf("missing domain in email")
        }
        return nil
    })),
)
```

### Struct Validation

```go
type User struct {
    Name string
    Age  int
}

actor.AttemptsTo(
    ensure.That(answerable.ValueOf(user), expectations.Satisfies("has valid user data", func(actual User) error {
        if actual.Name == "" {
            return fmt.Errorf("name is empty")
        }
        if actual.Age < 18 {
            return fmt.Errorf("age %d is too young (minimum 18)", actual.Age)
        }
        if actual.Age > 100 {
            return fmt.Errorf("age %d is unrealistic (maximum 100)", actual.Age)
        }
        return nil
    })),
)
```

## Advanced Usage with go-cmp

### Struct Comparison

```go
import "github.com/google/go-cmp/cmp"

type User struct {
    Name string
    Age  int
    Tags []string
}

expected := User{Name: "Alice", Age: 25, Tags: []string{"admin", "user"}}
actual := User{Name: "Alice", Age: 25, Tags: []string{"admin", "user"}}

actor.AttemptsTo(
    ensure.That(answerable.ValueOf(actual), expectations.Satisfies("matches expected user structure", func(actual User) error {
        if diff := cmp.Diff(expected, actual); diff != "" {
            return fmt.Errorf("user struct mismatch (-expected +actual):\n%s", diff)
        }
        return nil
    })),
)
```

### Comparison with Options

```go
import (
    "github.com/google/go-cmp/cmp"
    "github.com/google/go-cmp/cmp/cmpopts"
)

type TimestampedUser struct {
    ID        int
    Name      string
    CreatedAt time.Time
    UpdatedAt time.Time
}

actor.AttemptsTo(
    ensure.That(answerable.ValueOf(actual), expectations.Satisfies("matches user ignoring timestamps", func(actual TimestampedUser) error {
        if diff := cmp.Diff(expected, actual,
            cmpopts.IgnoreFields(TimestampedUser{}, "CreatedAt", "UpdatedAt"),
            cmpopts.EquateEmpty()); diff != "" {
            return fmt.Errorf("user struct mismatch (-expected +actual):\n%s", diff)
        }
        return nil
    })),
)
```

### Slice Comparison with Sorting

```go
type Item struct {
    ID   int
    Name string
}

expected := []Item{
    {ID: 2, Name: "item2"},
    {ID: 1, Name: "item1"},
}

actual := []Item{
    {ID: 1, Name: "item1"},
    {ID: 2, Name: "item2"},
}

actor.AttemptsTo(
    ensure.That(answerable.ValueOf(actual), expectations.Satisfies("matches items ignoring order", func(actual []Item) error {
        if diff := cmp.Diff(expected, actual,
            cmpopts.SortSlices(func(a, b Item) bool { return a.ID < b.ID })); diff != "" {
            return fmt.Errorf("items slice mismatch (-expected +actual):\n%s", diff)
        }
        return nil
    })),
)
```

## Complex Business Logic Validation

### Order Validation

```go
type Order struct {
    ID        string
    Amount    float64
    Currency  string
    Status    string
    CreatedAt time.Time
    Items     []OrderItem
}

type OrderItem struct {
    ProductID string
    Quantity  int
    Price     float64
}

actor.AttemptsTo(
    ensure.That(answerable.ValueOf(order), expectations.Satisfies("has valid order data", func(actual Order) error {
        if !strings.HasPrefix(actual.ID, "ORD-") {
            return fmt.Errorf("order ID must start with ORD-, got %s", actual.ID)
        }
        if actual.Amount <= 0 {
            return fmt.Errorf("order amount must be positive, got %f", actual.Amount)
        }

        validCurrencies := []string{"USD", "EUR", "GBP"}
        currencyValid := false
        for _, currency := range validCurrencies {
            if actual.Currency == currency {
                currencyValid = true
                break
            }
        }
        if !currencyValid {
            return fmt.Errorf("invalid currency %s, valid options: %v", actual.Currency, validCurrencies)
        }

        if len(actual.Items) == 0 {
            return fmt.Errorf("order must have at least one item")
        }

        var calculatedTotal float64
        for _, item := range actual.Items {
            if item.Quantity <= 0 {
                return fmt.Errorf("item quantity must be positive, got %d for product %s", item.Quantity, item.ProductID)
            }
            if item.Price <= 0 {
                return fmt.Errorf("item price must be positive, got %f for product %s", item.Price, item.ProductID)
            }
            calculatedTotal += float64(item.Quantity) * item.Price
        }

        if diff := actual.Amount - calculatedTotal; diff > 0.01 || diff < -0.01 {
            return fmt.Errorf("order amount %f doesn't match calculated total %f", actual.Amount, calculatedTotal)
        }

        return nil
    })),
)
```

## Error Messages

The description you provide to `Satisfies` appears in test failure messages:

```
actor ensures that 42 (int) is positive number failed: assertion failed for '42 (int)': expected positive value, but got -5
```

## Best Practices

1. **Use descriptive descriptions** — Make it clear what the validation is checking
2. **Provide detailed error messages** — Include the actual and expected values
3. **Keep validation focused** — Each `Satisfies` should check one logical condition
4. **Use go-cmp for struct comparisons** — Leverage go-cmp for complex struct comparisons
5. **Handle edge cases** — Consider nil values, empty collections, and boundary conditions

## Integration with Existing Expectations

`Satisfies` works seamlessly with built-in expectations:

```go
actor.AttemptsTo(
    ensure.That(api.LastResponseStatus{}, expectations.Equals(200)),
    ensure.That(answerable.ValueOf(responseData), expectations.Satisfies("has valid response structure", func(actual ResponseData) error {
        // Custom validation logic
        return nil
    })),
)
```

## Type Safety

`Satisfies` is type-safe via generics:

```go
// This will not compile - wrong type
expectations.Satisfies("is positive", func(actual string) error { ... })

// Correct
expectations.Satisfies("is positive", func(actual int) error { ... })
```

## Running the Examples

Working examples are available in `examples/satisfies_demo_test.go` in the lib repo:

```bash
go test ./examples -v -run TestSatisfies
```
