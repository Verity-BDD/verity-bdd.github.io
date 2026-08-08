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
    api.SendPostRequest("/posts").WithBody(map[string]any{ ... }),
    ...
)
```

`AttemptsTo` executes activities in order. The HTTP interactions (`SendGetRequest`, `SendPostRequest`, etc.) use the actor's `CallAnApiAt` ability to send requests.

Available HTTP interactions:

```go
api.SendGetRequest("/path")
api.SendPostRequest("/path").WithBody(body)
api.SendPutRequest("/path").WithBody(body)
api.SendPatchRequest("/path").WithBody(body)
api.SendDeleteRequest("/path")

// Add headers to any request:
api.SendGetRequest("/path").WithHeader("Authorization", "Bearer token")
api.SendPostRequest("/path").
    WithHeader("X-Request-ID", "123").
    WithBody(body)
```

### 4. Assert with questions and expectations

```go
ensure.That(api.LastResponseStatus{}, expectations.Equals(201)),
ensure.That(api.LastResponseBody{}, expectations.ContainsSubstring("My First Post")),
```

`ensure.That` takes a **question** (something to retrieve) and an **expectation** (the condition to check).

Built-in API questions:

```go
api.LastResponseStatus{}                // HTTP status code (int)
api.LastResponseBody{}                  // Response body (string)
api.NewResponseHeader("content-type")   // Response header value (string)
api.NewJSONPath("data.user.name")       // JSONPath result (any)
api.LastResponseBodyAsJSON[T]()          // Deserialise body into T
```

`NewJSONPath` uses dot-separated object keys and numeric array indexes. A `*` array segment returns `[]any`; JSON numbers decode with the standard `encoding/json` representation (normally `float64`). The prebuilt `api.LastResponseStatusQ` and `api.LastResponseBodyQ` are equivalent to the empty struct questions above.

`api.ResponseTime{}` and the equivalent `api.ResponseTimeQ` currently always answer `0`; request timing is not implemented yet, so do not use either for performance assertions.

### Advanced request entry points

Use `api.Using(client)` when the actor needs a configured `*http.Client` (for example, custom transport, cookies, redirects, or timeouts). Because it has no base URL, send absolute request URLs:

```go
client := &http.Client{Timeout: 10 * time.Second}
author := test.ActorCalled("Author").WhoCan(api.Using(client))

request, err := api.NewRequestBuilder(http.MethodGet, "https://api.example.org/posts").
    WithHeader("Accept", "application/json").
    Build()
if err != nil {
    t.Fatal(err)
}

author.AttemptsTo(api.SendRequest(request))
```

`NewRequestBuilder` also supports `WithHeaders`, `WithBody(io.Reader)`, `With(data)`, and `WithJSONBody(data)`. `WithJSONBody` returns an error; `With` is fluent and attempts JSON encoding for other values. You can also build a standard `*http.Request` directly and pass it to `SendRequest`.

## The complete test

```go title="post_api_test.go"
package myapp_test

import (
    "testing"

    verity "github.com/verity-bdd/verity-bdd"
    "github.com/verity-bdd/verity-bdd/verity_abilities/api"
    expectations "github.com/verity-bdd/verity-bdd/verity_expectations"
    "github.com/verity-bdd/verity-bdd/verity_expectations/ensure"
)

func TestCreatePost(t *testing.T) {
    test := verity.NewVerityTest(t, verity.Scene{})

    author := test.ActorCalled("Author").WhoCan(
        api.CallAnApiAt("https://jsonplaceholder.typicode.com"),
    )

    author.AttemptsTo(
        api.SendPostRequest("/posts").WithBody(map[string]any{
            "title":  "My First Post",
            "body":   "Hello from Verity BDD",
            "userId": 1,
        }),
        ensure.That(api.LastResponseStatus{}, expectations.Equals(201)),
        ensure.That(api.LastResponseBody{}, expectations.ContainsSubstring("My First Post")),
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
  ✅ Author sends POST request to /posts (0.18s)
  ✅ Author ensures that the last response status code equals 201 (0.00s)
  ✅ Author ensures that the last response body contains 'My First Post' (0.00s)
✅ TestCreatePost: PASSED (0.19s)
```

`OnStepStart` tracks the active step but does not print a start line; the console reporter writes each step only when it finishes. Durations vary by run.

For Allure reporting, see the [Reporting guide](/en/get_started/20_reporting/).
