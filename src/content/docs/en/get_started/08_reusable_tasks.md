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

