---
title: Ability Examples
description: Build a custom stateful ability with current Verity BDD APIs
---

An ability is an object stored on an actor. Activities use it to change the system or the ability's state; questions use it to retrieve values for assertions.

`verity.Ability` is currently an empty interface, so every Go type satisfies it. Defining a domain-specific interface is still useful because `verity.AbilityOf[T](actor)` returns that exact interface and keeps activities decoupled from the implementation. Embedding `verity.Ability` is an optional marker convention, not a compiler-enforced requirement.

This example implements a small in-memory key/value store. The same structure works for database clients, file systems, queues, WebSocket clients, and other integrations.

## Define and implement the ability

```go title="keystore/ability.go"
package keystore

import (
    "fmt"
    "sync"

    verity "github.com/verity-bdd/verity-bdd"
)

// Ability describes only the operations activities and questions need.
type Ability interface {
    verity.Ability // optional marker
    Put(key, value string)
    Get(key string) (string, error)
}

type memoryStore struct {
    mu     sync.RWMutex
    values map[string]string
}

func InMemory() Ability {
    return &memoryStore{values: make(map[string]string)}
}

func (s *memoryStore) Put(key, value string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.values[key] = value
}

func (s *memoryStore) Get(key string) (string, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    value, ok := s.values[key]
    if !ok {
        return "", fmt.Errorf("key %q not found", key)
    }
    return value, nil
}
```

Create a separate stateful instance for each actor unless shared state is intentional:

```go
alice := test.ActorCalled("Alice").WhoCan(keystore.InMemory())
bob := test.ActorCalled("Bob").WhoCan(keystore.InMemory())
```

`WhoCan` appends abilities; it does not enforce one ability per type. `AbilityOf[T]` returns the first stored ability assignable to `T`. Request the interface itself, not a pointer to it:

```go
store, err := verity.AbilityOf[keystore.Ability](actor) // correct
// verity.AbilityOf[*keystore.Ability](actor)           // pointer-to-interface: wrong type
```

## Add an activity

An activity implements `PerformAs(context.Context, verity.Actor) error`, `Description() string`, and `FailureMode() verity.FailureMode`:

```go title="keystore/activities.go"
package keystore

import (
    "context"
    "fmt"

    verity "github.com/verity-bdd/verity-bdd"
)

type putActivity struct {
    key   string
    value string
}

func Put(key, value string) verity.Activity {
    return &putActivity{key: key, value: value}
}

func (a *putActivity) PerformAs(_ context.Context, actor verity.Actor) error {
    store, err := verity.AbilityOf[Ability](actor)
    if err != nil {
        return fmt.Errorf("put value: %w", err)
    }
    store.Put(a.key, a.value)
    return nil
}

func (a *putActivity) Description() string {
    return fmt.Sprintf("#actor stores value under %q", a.key)
}

func (*putActivity) FailureMode() verity.FailureMode {
    return verity.Critical()
}
```

`verity.Do` is the shorter option when a named activity type is unnecessary. Its callback also receives the context and actor:

```go
verity.Do("#actor clears the remote cache", func(ctx context.Context, actor verity.Actor) error {
    cache, err := verity.AbilityOf[CacheAbility](actor)
    if err != nil {
        return err
    }
    return cache.Clear(ctx)
})
```

## Add a question

Use `verity.QuestionAbout` for a dynamic value. The callback is evaluated when an assertion asks the question:

```go title="keystore/questions.go"
package keystore

import (
    "context"
    "fmt"

    verity "github.com/verity-bdd/verity-bdd"
)

func Value(key string) verity.Question[string] {
    return verity.QuestionAbout(
        fmt.Sprintf("value stored under %q", key),
        func(_ context.Context, actor verity.Actor) (string, error) {
            store, err := verity.AbilityOf[Ability](actor)
            if err != nil {
                return "", fmt.Errorf("read value: %w", err)
            }
            return store.Get(key)
        },
    )
}
```

Use `verity_answerable.ValueOf(value)` only for an already-known static value. It does not create dynamic ability-backed questions.

## Use the ability

```go title="keystore_test.go"
package keystore_test

import (
    "testing"

    verity "github.com/verity-bdd/verity-bdd"
    expectations "github.com/verity-bdd/verity-bdd/verity_expectations"
    "github.com/verity-bdd/verity-bdd/verity_expectations/ensure"

    "myproject/keystore"
)

func TestStoreValue(t *testing.T) {
    test := verity.NewVerityTest(t, verity.Scene{})
    actor := test.ActorCalled("Alice").WhoCan(keystore.InMemory())

    actor.AttemptsTo(
        keystore.Put("status", "ready"),
        ensure.That(keystore.Value("status"), expectations.Equals("ready")),
        ensure.That(keystore.Value("status"), expectations.ContainsSubstring("read")),
    )
}
```

`NewVerityTest` registers cleanup with the supplied `TestContext`, so an ordinary Go test does not need to call `Shutdown` manually.

## Failure modes

Every activity chooses its failure behavior through `FailureMode()`:

| Return value | Behavior in `AttemptsTo` |
|---|---|
| `verity.Critical()` (or `verity.FailFast`) | Report the error, call `FailNow`, and stop the sequence |
| `verity.NonCritical()` (or `verity.ErrorButContinue`) | Report the error with `Errorf` and continue |
| `verity.Optional()` (or `verity.Ignore`) | Log the ignored error and continue |

The semantic functions return failure-mode values; they are not fluent modifiers. For example, a custom best-effort cleanup activity can return `verity.Optional()` from its `FailureMode` method. `verity.Do` and built-in request activities are fail-fast, while a plain `ensure.That` is non-critical and `.After(...)` is fail-fast.

Choose the mode in the activity implementation rather than expecting `AttemptsTo` to return an error: `AttemptsTo` returns no value and reports failures through the test's `TestContext`.
