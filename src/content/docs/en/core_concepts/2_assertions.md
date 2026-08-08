---
title: Assertions and expectations
description: How assertions and expectations work
sidebar:
  order: 2
---

## Assertions and expectations

Verity BDD helps you model test scenarios from the perspective of [actors performing activities](/en/core_concepts/1_screenplay/#actors) to accomplish their goals.
Assertions follow this same consistent approach, expressed using the `ensure.That` activity.

### The anatomy of a Verity BDD assertion

`ensure.That` accepts two arguments:
- the **actual value** — a `Question[T]` to be evaluated in the context of the given actor,
- an **`Expectation[T]`** — the condition to be met by the actual value.

```go
import (
    verity "github.com/verity-bdd/verity-bdd"
    answerable "github.com/verity-bdd/verity-bdd/verity_answerable"
    expectations "github.com/verity-bdd/verity-bdd/verity_expectations"
    "github.com/verity-bdd/verity-bdd/verity_expectations/ensure"
)

actor.AttemptsTo(
    ensure.That(answerable.ValueOf("Hello world"), expectations.ContainsSubstring("Hello")),
    //          actual value ----^                  ^---- expectation
)
```

The available built-in expectations are:

| Expectation | Description |
|---|---|
| `expectations.Equals(expected)` | Any value deeply equals `expected` (`reflect.DeepEqual`) |
| `expectations.ContainsSubstring(substr)` | String contains `substr` |
| `expectations.ContainsKey(key)` | Map contains the given key |
| `expectations.Includes(value)` | Slice contains an element deeply equal to `value` |
| `expectations.IsGreaterThan(n)` | Numeric value is greater than `n` |
| `expectations.IsLessThan(n)` | Numeric value is less than `n` |
| `expectations.IsEmpty[T]()` | String, slice, array, or map has length zero |
| `expectations.ArrayLengthEquals[T](n)` | Array, slice, or string has length `n` |
| `expectations.Not(inner)` | Negates any expectation |
| `expectations.Satisfies(desc, fn)` | Custom expectation via a validation function |

### Static and dynamic questions

The actual value in `ensure.That` is always a `Question[T]`. Verity BDD provides two ways to create one:

**Static values** — wrap any value with `answerable.ValueOf`:

```go
import (
    answerable "github.com/verity-bdd/verity-bdd/verity_answerable"
    expectations "github.com/verity-bdd/verity-bdd/verity_expectations"
    "github.com/verity-bdd/verity-bdd/verity_expectations/ensure"
)

actor.AttemptsTo(
    ensure.That(answerable.ValueOf(42), expectations.Equals(42)),
    ensure.That(answerable.ValueOf("hello@example.com"), expectations.ContainsSubstring("@")),
    ensure.That(answerable.ValueOf(true), expectations.Equals(true)),
)
```

**Dynamic values** — questions evaluated at assertion time, such as API response questions:

```go
import (
    "github.com/verity-bdd/verity-bdd/verity_abilities/api"
    expectations "github.com/verity-bdd/verity-bdd/verity_expectations"
    "github.com/verity-bdd/verity-bdd/verity_expectations/ensure"
)

apisitt.AttemptsTo(
    api.SendGetRequest("/products"),
    ensure.That(api.LastResponseStatus{}, expectations.Equals(200)),
    ensure.That(api.LastResponseBody{}, expectations.ContainsSubstring("Apples")),
)
```

`api.LastResponseStatus{}` and `api.LastResponseBody{}` are questions that retrieve the HTTP status code
and response body from the actor's last API interaction, evaluated at assertion time.

### Reusable assertions

Since `ensure.That` is just an activity, it can be composed into reusable tasks.
This is what makes Verity BDD assertions especially powerful — the same assertion logic
can be reused across different test scenarios.

Consider a task that verifies a URL returns 200 OK:

```go title="tasks.go"
import (
    verity "github.com/verity-bdd/verity-bdd"
    "github.com/verity-bdd/verity-bdd/verity_abilities/api"
    expectations "github.com/verity-bdd/verity-bdd/verity_expectations"
    "github.com/verity-bdd/verity-bdd/verity_expectations/ensure"
)

func CheckURL(path string) verity.Activity {
    return verity.TaskWhere("check "+path+" returns 200",
        api.SendGetRequest(path),
        ensure.That(api.LastResponseStatus{}, expectations.Equals(200)),
    )
}
```

You can then use this task to build a simple link checker:

```go title="link_checker_test.go"
func TestLinkChecker(t *testing.T) {
    ctx := context.Background()
    test := verity.NewVerityTestWithContext(ctx, t)

    apisitt := test.ActorCalled("Apisitt").
        WhoCan(api.CallAnApiAt("https://api.example.org/"))

    for _, path := range []string{"/products", "/users", "/orders"} {
        apisitt.AttemptsTo(CheckURL(path))
    }
}
```

### Custom expectations

When built-in expectations are not enough, use `expectations.Satisfies` to write a custom validation function.
The function receives the actual value and returns an error if the expectation is not met:

```go
import (
    "fmt"
    "strings"

    answerable "github.com/verity-bdd/verity-bdd/verity_answerable"
    expectations "github.com/verity-bdd/verity-bdd/verity_expectations"
    "github.com/verity-bdd/verity-bdd/verity_expectations/ensure"
)

actor.AttemptsTo(
    ensure.That(
        answerable.ValueOf("hello@example.com"),
        expectations.Satisfies("is a valid email", func(actual string) error {
            if !strings.Contains(actual, "@") {
                return fmt.Errorf("missing @ symbol: %q", actual)
            }
            if !strings.Contains(actual, ".") {
                return fmt.Errorf("missing domain in email: %q", actual)
            }
            return nil
        }),
    ),
)
```

`Satisfies` works with any type, including structs:

```go
type User struct {
    Name string
    Age  int
}

actor.AttemptsTo(
    ensure.That(
        answerable.ValueOf(User{Name: "Alice", Age: 25}),
        expectations.Satisfies("is a valid adult user", func(u User) error {
            if u.Name == "" {
                return fmt.Errorf("name is empty")
            }
            if u.Age < 18 {
                return fmt.Errorf("age %d is below 18", u.Age)
            }
            return nil
        }),
    ),
)
```

### Struct comparison with `go-cmp`

For comparing complex structs, `Satisfies` pairs well with [`github.com/google/go-cmp`](https://pkg.go.dev/github.com/google/go-cmp/cmp):

```go
import (
    "fmt"

    "github.com/google/go-cmp/cmp"
    "github.com/google/go-cmp/cmp/cmpopts"

    answerable "github.com/verity-bdd/verity-bdd/verity_answerable"
    expectations "github.com/verity-bdd/verity-bdd/verity_expectations"
    "github.com/verity-bdd/verity-bdd/verity_expectations/ensure"
)

type User struct {
    Name string
    Age  int
    Tags []string
}

expected := User{Name: "Alice", Age: 25, Tags: []string{"admin"}}

actor.AttemptsTo(
    ensure.That(
        answerable.ValueOf(actualUser),
        expectations.Satisfies("matches expected user", func(actual User) error {
            if diff := cmp.Diff(expected, actual); diff != "" {
                return fmt.Errorf("user mismatch (-expected +actual):\n%s", diff)
            }
            return nil
        }),
    ),
)
```

You can use `cmpopts` to ignore specific fields — handy for timestamps or server-generated IDs:

```go
expectations.Satisfies("matches user ignoring timestamps", func(actual User) error {
    if diff := cmp.Diff(expected, actual,
        cmpopts.IgnoreFields(User{}, "CreatedAt", "UpdatedAt"),
    ); diff != "" {
        return fmt.Errorf("user mismatch (-expected +actual):\n%s", diff)
    }
    return nil
})
```

### Negating expectations

Use `expectations.Not` to negate any expectation:

```go
actor.AttemptsTo(
    ensure.That(api.LastResponseStatus{}, expectations.Not(expectations.Equals(404))),
    ensure.That(api.LastResponseBody{}, expectations.Not(expectations.ContainsSubstring("error"))),
)
```

### Custom questions

You can define your own `Question[T]` using `verity.QuestionAbout` for values that need to be
retrieved dynamically at assertion time:

```go
import (
    "context"
    "encoding/json"
    "fmt"

    verity "github.com/verity-bdd/verity-bdd"
    "github.com/verity-bdd/verity-bdd/verity_abilities/api"
    expectations "github.com/verity-bdd/verity-bdd/verity_expectations"
    "github.com/verity-bdd/verity-bdd/verity_expectations/ensure"
)

type Product struct {
    Name  string `json:"name"`
    Price string `json:"price"`
}

func FirstProductName() verity.Question[string] {
    return verity.QuestionAbout("name of the first product", func(ctx context.Context, actor verity.Actor) (string, error) {
        body, err := api.LastResponseBody{}.AnsweredBy(ctx, actor)
        if err != nil {
            return "", err
        }
        var products []Product
        if err := json.Unmarshal([]byte(body), &products); err != nil {
            return "", err
        }
        if len(products) == 0 {
            return "", fmt.Errorf("no products in response")
        }
        return products[0].Name, nil
    })
}

// AssertFirstProductIsApples shows the question at a complete call site.
func AssertFirstProductIsApples(apisitt verity.Actor) {
    apisitt.AttemptsTo(
        api.SendGetRequest("/products"),
        ensure.That(FirstProductName(), expectations.Equals("Apples")),
    )
}
```

### Delayed and polling assertions

`ensure.That(question, expectation).After(duration)` waits once, then answers the question and evaluates the expectation once. It is a deliberate delay, not a polling mechanism:

```go
actor.AttemptsTo(
    ensure.That(Status(), expectations.Equals("ready")).After(2*time.Second),
)
```

The delayed activity is fail-fast. Cancellation of the actor's context interrupts the delay.

For polling, use `wait.Until`. It checks immediately, then repeats until the expectation passes, the context is cancelled, or the timeout expires. Defaults are a 5-second timeout and a 500-millisecond interval:

```go
import "github.com/verity-bdd/verity-bdd/verity_abilities/wait"

actor.AttemptsTo(
    wait.Until(Status(), expectations.Equals("ready")).
        For(30*time.Second).
        CheckingEvery(time.Second),
)
```

`wait.UntilReceived(events).For(10*time.Second)` instead waits for one value from a receive-only channel. It fails if the channel closes before a value arrives, the timeout expires, or the context is cancelled. Both wait activities are fail-fast.

### Expectations whose expected value is a question

Dynamic expectation factories resolve another question while evaluating the actual answer:

```go
actor.AttemptsTo(
    ensure.That(CurrentUserName(), expectations.EqualsAnswerTo(ExpectedUserName())),
    ensure.That(ResponseText(), expectations.ContainsSubstringAnswerTo(SearchTerm())),
    ensure.That(ResponseMap(), expectations.ContainsKeyAnswerTo(RequiredKey())),
    ensure.That(Items(), expectations.ArrayLengthEqualsAnswerTo[[]Item](ExpectedCount())),
)
```

The complete set is `EqualsAnswerTo`, `ContainsSubstringAnswerTo`, `ContainsKeyAnswerTo`, `ArrayLengthEqualsAnswerTo`, `IsGreaterThanAnswerTo`, and `IsLessThanAnswerTo`. The numeric variants use `Question[any]` because they accept the supported numeric types. `SatisfiesAnswer` is the custom form; its callback receives `context.Context`, the current `verity.Actor`, and the actual value:

```go
expectations.SatisfiesAnswer("is visible to the current actor",
    func(ctx context.Context, actor verity.Actor, actual Resource) error {
        return checkVisibility(ctx, actor, actual)
    },
)
```

The expected-value question is answered each time the expectation is evaluated. This matters inside `wait.Until`, where both sides can change on every poll. If resolving the expected question fails, the assertion reports that question-resolution error rather than treating it as a normal mismatch.
