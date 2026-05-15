---
title: Writing Your First Test
description: Write your first Screenplay-style API test with Verity BDD
sidebar:
  order: 2
---

This guide walks you through writing a complete API test using the Screenplay Pattern.
We'll use [JSONPlaceholder](https://jsonplaceholder.typicode.com) — a free public API — as the system under test.

## Step by step

### 1. Create a VerityTest

```go
test := verity.NewVerityTest(t, verity.Scene{})
```

`VerityTest` is the entry point for each test. It manages actors and automatically cleans up after the test finishes.

### 2. Create an actor with an ability

```go
author := test.ActorCalled("Author").WhoCan(
    api.CallAnApiAt("https://jsonplaceholder.typicode.com"),
)
```

- `ActorCalled("Author")` creates an actor named "Author".
- `WhoCan(...)` gives the actor the ability to make HTTP calls to the given base URL.

The base URL is stored in the ability — individual interactions only specify the path.

### 3. Perform interactions

```go
author.AttemptsTo(
    api.SendPostRequest("/posts").With(map[string]any{ ... }),
    ...
)
```

`AttemptsTo` executes activities in order. The HTTP interactions (`SendGetRequest`, `SendPostRequest`, etc.) use the actor's `CallAnApiAt` ability to send requests.

Available HTTP interactions:

```go
api.SendGetRequest("/path")
api.SendPostRequest("/path").With(body)
api.SendPutRequest("/path").With(body)
api.SendDeleteRequest("/path")

// Add headers to any request:
api.SendGetRequest("/path").WithHeader("Authorization", "Bearer token")
api.SendPostRequest("/path").
    WithHeader("X-Request-ID", "123").
    With(body)
```

### 4. Assert with questions and expectations

```go
ensure.That(api.LastResponseStatus{}, expectations.Equals(201)),
ensure.That(api.LastResponseBody{}, expectations.Contains("My First Post")),
```

`ensure.That` takes a **question** (something to retrieve) and an **expectation** (the condition to check).

Built-in API questions:

```go
api.LastResponseStatus{}                // HTTP status code (int)
api.LastResponseBody{}                  // Response body (string)
api.NewResponseHeader("content-type")   // Response header value (string)
api.NewJSONPath("data.user.name")       // JSONPath expression result (string)
api.ResponseTime{}                      // Response time in milliseconds (int64)
api.NewResponseBodyAsJSON[T]()          // Deserialise body into struct T
```

## The complete test

```go title="post_api_test.go"
package myapp_test

import (
    "testing"

    verity "github.com/nchursin/verity-bdd"
    "github.com/nchursin/verity-bdd/verity_abilities/api"
    expectations "github.com/nchursin/verity-bdd/verity_expectations"
    "github.com/nchursin/verity-bdd/verity_expectations/ensure"
)

func TestCreatePost(t *testing.T) {
    test := verity.NewVerityTest(t, verity.Scene{})

    author := test.ActorCalled("Author").WhoCan(
        api.CallAnApiAt("https://jsonplaceholder.typicode.com"),
    )

    author.AttemptsTo(
        api.SendPostRequest("/posts").With(map[string]any{
            "title":  "My First Post",
            "body":   "Hello from Verity BDD",
            "userId": 1,
        }),
        ensure.That(api.LastResponseStatus{}, expectations.Equals(201)),
        ensure.That(api.LastResponseBody{}, expectations.Contains("My First Post")),
    )
}
```

Run it with:

```bash
go test ./...
```


## Console output

By default, Verity BDD prints a step-by-step execution log to the console:

```
🚀 Starting: TestCreatePost
  🔄 Author sends POST request to /posts
  ✅ Author sends POST request to /posts (0.18s)
  🔄 Ensures that the last response status code equals 201
  ✅ Ensures that the last response status code equals 201 (0.00s)
  🔄 Ensures that the last response body contains "My First Post"
  ✅ Ensures that the last response body contains "My First Post" (0.00s)
✅ TestCreatePost: PASSED (0.19s)
```

For Allure reporting, see the [Reporting guide](/en/guides/3_reporting/).
