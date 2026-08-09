---
title: Reusable Tasks
description: Extract repeated sequences of interactions into named tasks using `verity.TaskWhere`
sidebar:
  order: 8
---

Extract repeated sequences of interactions into named tasks using `verity.TaskWhere`:

```go title="tasks.go"
package myapp_test

import (
    verity "github.com/verity-bdd/verity-bdd"
    "github.com/verity-bdd/verity-bdd/verity_abilities/api"
    expectations "github.com/verity-bdd/verity-bdd/verity_expectations"
    "github.com/verity-bdd/verity-bdd/verity_expectations/ensure"
)

type Post struct {
    Title  string `json:"title"`
    Body   string `json:"body"`
    UserID int    `json:"userId"`
}

func PublishPost(post Post) verity.Activity {
    return verity.TaskWhere("#actor publishes a post",
        api.SendPostRequest("/posts").WithBody(post),
        ensure.That(api.LastResponseStatus{}, expectations.Equals(201)),
    )
}
```

`TaskWhere` executes children sequentially and stops at the first child error. The enclosing task itself is fail-fast, regardless of the failed child's own failure mode. For example, an `ensure.That` attempted directly can report an error and let a later top-level activity run, but the same assertion inside this task returns an error to the task, prevents remaining children from running, and makes `AttemptsTo` stop at the task boundary. Reporters receive the enclosing task activity and its nested child activities, preserving both business-level and detailed reporting.

```go title="post_api_test.go"
func TestPublishPost(t *testing.T) {
    test := verity.NewVerityTest(t, verity.Scene{})

    author := test.ActorCalled("Author").WhoCan(
        api.CallAnApiAt("https://jsonplaceholder.typicode.com"),
    )

    author.AttemptsTo(
        PublishPost(Post{Title: "My First Post", Body: "Hello!", UserID: 1}),
    )
}
```

