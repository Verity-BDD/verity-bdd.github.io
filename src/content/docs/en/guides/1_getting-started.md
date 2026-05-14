---
title: Getting Started
description: Install Verity BDD and write your first test
sidebar:
  order: 1
---

<Aside type="caution">
  This project is still at version 0.x.x. No backwards compatibility is guaranteed for any changes. The plan is to go v1.x.x in Summer 2026.
</Aside>

A Go implementation of the Screenplay Pattern for acceptance testing, focused on API testing capabilities.

## Overview

Verity-BDD brings the power of the Screenplay Pattern to Go testing, providing:

- **Actor-centric testing** — Tests describe what actors do, not how they do it
- **Reusable components** — Build a library of reusable tasks and interactions
- **Clear domain language** — Tests that read like business requirements
- **Modular design** — Use only what you need for your testing scenarios
- **Framework agnostic** — Works with any Go test runner

## Installation

```bash
go get github.com/nchursin/verity-bdd
```

## Basic Example

```go
package main

import (
    "testing"

    "github.com/nchursin/verity-bdd/verity_abilities/api"
    expectations "github.com/nchursin/verity-bdd/verity_expectations"
    "github.com/nchursin/verity-bdd/verity_expectations/ensure"
    verity "github.com/nchursin/verity-bdd"
)

func TestAPI(t *testing.T) {
    test := verity.NewVerityTest(t, verity.Scene{})

    // Create an actor with API ability
    actor := test.ActorCalled("APITester").WhoCan(
        api.CallAnApiAt("https://jsonplaceholder.typicode.com"),
    )

    // Define test data
    newPost := map[string]interface{}{
        "title":  "Test Post",
        "body":   "This is a test post",
        "userId": 1,
    }

    // Test the API
    actor.AttemptsTo(
        api.SendPostRequest("/posts").
            WithBody(newPost),
        ensure.That(api.LastResponseStatus{}, expectations.Equals(201)),
        ensure.That(api.LastResponseBody{}, expectations.Contains("Test Post")),
    )
}
```

## Core Concepts

### Actors

Actors represent people or systems interacting with your application:

```go
test := verity.NewVerityTest(t, verity.Scene{})

// Create an actor
actor := test.ActorCalled("John Doe")

// Give the actor abilities to interact with your system
actor = actor.WhoCan(api.CallAnApiAt("https://api.example.com"))
```

### Abilities

Abilities enable actors to interact with different interfaces:

```go
// HTTP API ability
apiAbility := api.CallAnApiAt("https://api.example.com")

// Actor with multiple abilities
actor := test.ActorCalled("TestUser").WhoCan(
    apiAbility,
    // ... other abilities
)
```

### Activities

Activities represent actions that actors perform.

#### Interactions (low-level actions)
```go
api.SendGetRequest("/users")
api.SendPostRequest("/posts").WithBody(postData)
api.SendPutRequest("/users/1").WithBody(updatedUser)
api.SendDeleteRequest("/posts/123")
```

#### Tasks (high-level business actions)
```go
createUserTask := core.Where(
    "creates a new user",
    core.Do("creates a new user", func(a core.Actor) error {
        req, err := api.NewRequestBuilder("POST", "/users").
            WithJSONBody(userData).
            Build()
        if err != nil {
            return err
        }
        return api.SendRequest(req).PerformAs(a)
    }),
    ensure.That(api.LastResponseStatus{}, expectations.Equals(201)),
)

actor.AttemptsTo(createUserTask)
```

### Questions

Questions retrieve information from the system:

```go
// Built-in questions
ensure.That(api.LastResponseStatus{}, expectations.Equals(200))
ensure.That(api.LastResponseBody{}, expectations.Contains("success"))
ensure.That(api.NewResponseHeader("content-type"), expectations.Contains("json"))

// Parse response as JSON struct
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

ensure.That(api.NewResponseBodyAsJSON[User](), expectations.Satisfies("has valid user", func(actual User) error {
    if actual.Name == "" {
        return fmt.Errorf("user name is empty")
    }
    return nil
}))

// JSONPath queries
ensure.That(api.NewJSONPath("name"), expectations.Contains("John"))
ensure.That(api.NewJSONPath("data.users.*.email"), expectations.Contains("@"))

// Response time
ensure.That(api.ResponseTime{}, expectations.IsLessThan(1000)) // milliseconds
```

### Assertions

Verify that expectations are met:

```go
ensure.That(question, expectations.Equals(expected))
ensure.That(question, expectations.Contains(substring))
ensure.That(question, expectations.IsEmpty())
ensure.That(question, expectations.ArrayLengthEquals(5))
ensure.That(question, expectations.IsGreaterThan(10))
ensure.That(question, expectations.ContainsKey("id"))

// Custom validation with Satisfies
ensure.That(answerable.ValueOf(value), expectations.Satisfies("custom description", func(actual T) error {
    // Your validation logic here
    return nil // or error with description
}))
```

## API Testing

### HTTP Requests

```go
// GET
actor.AttemptsTo(
    api.SendGetRequest("/posts"),
    ensure.That(api.LastResponseStatus{}, expectations.Equals(200)),
)

// POST with JSON body
actor.AttemptsTo(
    api.SendPostRequest("/posts").WithBody(map[string]interface{}{
        "title":  "New Post",
        "body":   "Post content",
        "userId": 1,
    }),
    ensure.That(api.LastResponseStatus{}, expectations.Equals(201)),
)

// PUT with headers
actor.AttemptsTo(
    api.SendPutRequest("/posts/1").
        WithHeader("Authorization", "Bearer token").
        WithBody(updatedData),
    ensure.That(api.LastResponseStatus{}, expectations.Equals(200)),
)

// DELETE
actor.AttemptsTo(
    api.SendDeleteRequest("/posts/1"),
    ensure.That(api.LastResponseStatus{}, expectations.Equals(200)),
)
```

## Reporting

### Console Reporting

The `NewVerityTest` API includes console reporting automatically:

```go
func TestAPITesting(t *testing.T) {
    test := verity.NewVerityTest(t, verity.Scene{})
    actor := test.ActorCalled("APITester").WhoCan(api.CallAnApiAt("https://jsonplaceholder.typicode.com"))

    actor.AttemptsTo(
        api.SendGetRequest("/posts"),
        ensure.That(api.LastResponseStatus{}, expectations.Equals(200)),
    )
}
```

Console output:
```
🚀 Starting: TestAPITesting
  🔄 Sends GET request to /posts
  ✅ Sends GET request to /posts (0.21s)
  🔄 Ensures that the last response status code equals 200
  ✅ Ensures that the last response status code equals 200 (0.00s)
✅ TestAPITesting: PASSED (0.26s)
```

### Allure Reporting

```go
import (
    "context"
    "github.com/nchursin/verity-bdd/verity_reporting/allure_reporter"
    verity "github.com/nchursin/verity-bdd"
)

func TestWithAllure(t *testing.T) {
    reporter := allure_reporter.NewAllureReporterWithDir("allure-results")

    test := verity.NewVerityTest(t, verity.Scene{
        Context:  context.Background(),
        Reporter: reporter,
    })

    actor := test.ActorCalled("Tester")
    actor.AttemptsTo(
        // your activities
    )
}
```

Generate a local HTML report after the test run:

```bash
allure serve allure-results
```

## Advanced Usage

### Task Composition

```go
setupTask := core.Where("setup test data", setupDataAction)
testTask := core.Where("run test scenario", testAction)
cleanupTask := core.Where("cleanup test data", cleanupAction)

actor.AttemptsTo(
    setupTask,
    testTask,
    cleanupTask,
)
```

### Multiple Actors

```go
test := verity.NewVerityTest(t, verity.Scene{})

admin := test.ActorCalled("Admin").WhoCan(api.CallAnApiAt(baseURL))
user := test.ActorCalled("RegularUser").WhoCan(api.CallAnApiAt(baseURL))

admin.AttemptsTo(createResourceTask)
user.AttemptsTo(accessResourceTask)
```

### Custom Interactions

```go
customInteraction := core.Do("performs custom action", func(actor core.Actor) error {
    // Your custom logic here
    return nil
})

actor.AttemptsTo(customInteraction)
```

### Custom Questions

```go
customQuestion := core.QuestionAbout[int]("custom value", func(actor core.Actor, ctx context.Context) (int, error) {
    return 42, nil
})

ensure.That(customQuestion, expectations.Equals(42))
```

## Architecture

### Package Structure

| Package | Description |
|---|---|
| `github.com/nchursin/verity-bdd` | Core Screenplay API, testing API, answerable helpers |
| `verity_abilities` | Default ability contracts |
| `verity_expectations` | Expectations and assertion helpers |
| `verity_expectations/ensure` | Ensure activities |
| `verity_reporting` | Reporting contracts and adapters |
| `verity_reporting/console_reporter` | Console reporter |
| `verity_reporting/allure_reporter` | Allure reporter |

### Design Principles

1. **Composable** — Build complex behaviors from simple components
2. **Reusable** — Create libraries of tasks and interactions
3. **Readable** — Tests that read like business specifications
4. **Extensible** — Add new abilities and integrations
5. **Type-safe** — Leverage Go's type system for safety
