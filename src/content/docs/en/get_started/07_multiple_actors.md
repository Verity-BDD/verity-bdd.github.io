---
title: Multiple Actors
description: Use multiple actors to model different roles in the same scenario
sidebar:
  order: 7
---

Use multiple actors to model different roles in the same scenario:

```go
func TestProductCatalogue(t *testing.T) {
    test := verity.NewVerityTest(t, verity.Scene{})

    manager := test.ActorCalled("Manager").WhoCan(
        api.CallAnApiAt("https://api.example.org"),
    )
    customer := test.ActorCalled("Customer").WhoCan(
        api.CallAnApiAt("https://api.example.org"),
    )

    manager.AttemptsTo(
        api.SendPostRequest("/products").WithBody(map[string]any{
            "name": "Apples", "price": "£2.50",
        }),
        ensure.That(api.LastResponseStatus{}, expectations.Equals(201)),
    )

    customer.AttemptsTo(
        api.SendGetRequest("/products"),
        ensure.That(api.LastResponseStatus{}, expectations.Equals(200)),
        ensure.That(api.LastResponseBody{}, expectations.ContainsSubstring("Apples")),
    )
}
```
