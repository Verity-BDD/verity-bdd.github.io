---
title: Notes Ability Examples
description: Examples using the TakeNotes ability for sharing data between steps
---

The `TakeNotes` ability lets an actor store and retrieve typed values during a test — useful for passing data between steps without shared variables.

## Basic Example

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

## How It Works

- `UsingEmptyNotepad()` — adds the note-taking ability to the actor
- `Using(NotepadWith(...))` — lets you pre-populate notes, e.g. actor name and role
- `TakeNoteOf(...).As("auth token")` — stores a value and adds a step to the report: `Nina takes note "auth token"`
- `Note[string]("auth token")` — reads a note with type checking
