---
title: Shared State Between Steps
description: Two ways to share and manage state between steps
sidebar:
  order: 11
---

There are 2 basic ways to share some data between steps:
1. Each actor has it's own instance for each ability, which means every ability carries over it's state between steps.
2. `take_notes` module, which is a basic ability to write some data to use between steps.

## Ability state
Whenever you assign an ability to an actor, a new instance should be created. This way each actor only uses it's own private ability. This in turn gives you an opportunity to share some data and state between the steps.  Simple assign ability field to some value and use it normally.

## `TakeNotes` ability
The `TakeNotes` ability basically uses the same mechanism, but it's built into the way framework works. It lets an actor store and retrieve typed values during a test — useful for passing data between steps without shared variables. And in the end all notes are attached to the test reports.

```go
package examples

import (
    "testing"

    "github.com/nchursin/verity-bdd/verity_abilities/take_notes"
    veritytesting "github.com/nchursin/verity-bdd"
)

func TestNotesExample(t *testing.T) {
    test := veritytesting.NewVerityTest(t, veritytesting.Scene{})
    actor := test.ActorCalled("Nina").WhoCan(take_notes.UsingEmptyNotepad())

    actor.AttemptsTo(
        take_notes.TakeNoteOf("Bearer abc123").As("auth token"),
    )

    token, ok := actor.AnswersTo(take_notes.Note[string]("auth token"))
    if !ok {
        t.Fatalf("note not found")
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
