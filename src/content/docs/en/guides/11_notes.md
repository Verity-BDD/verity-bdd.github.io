---
title: Shared State Between Steps
description: Two ways to share and manage state between steps
sidebar:
  order: 11
---

There are 2 basic ways to share some data between steps:
1. An ability can carry state between an actor's steps. The caller or a default-ability factory owns instance isolation; Verity stores the supplied reference unchanged.
2. The `take_notes` package provides a built-in ability for storing values between steps.

## Ability state
Create a separate stateful ability instance for each actor. Passing the same instance to multiple actors intentionally shares its state, so the ability itself must then provide any required concurrency safety. Activities and questions can update and read the selected instance between steps.

## `TakeNotes` ability
The `TakeNotes` ability uses the same mechanism but is integrated with the framework. It lets an actor store and retrieve typed values during a test. Non-empty notes are serialised into one test-level `"notes"` attachment when the test shuts down.

```go
package examples

import (
    "testing"

    "github.com/verity-bdd/verity-bdd/verity_abilities/take_notes"
    veritytesting "github.com/verity-bdd/verity-bdd"
)

func TestNotesExample(t *testing.T) {
    test := veritytesting.NewVerityTest(t, veritytesting.Scene{})
    actor := test.ActorCalled("Nina").WhoCan(take_notes.UsingEmptyNotepad())

    actor.AttemptsTo(
        take_notes.TakeNoteOf("Bearer abc123").As("auth token"),
    )

    token, err := take_notes.Note[string]("auth token").AnsweredBy(test.Context(), actor)
    if err != nil {
        t.Fatal(err)
    }
    if token != "Bearer abc123" {
        t.Fatalf("unexpected token: %s", token)
    }
}
```

### How It Works

- `UsingEmptyNotepad()` — adds the note-taking ability to the actor
- `Using(NotepadWith(...))` — lets you pre-populate notes, e.g. actor name and role
- `TakeNoteOf(...).As("auth token")` — stores a value and adds a step to the report: `Nina takes note "auth token"`
- `Note[string]("auth token")` — reads a note with type checking
