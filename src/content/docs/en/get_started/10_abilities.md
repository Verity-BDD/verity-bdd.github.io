---
title: Creating Custom Abilities
description: Learn how to build your own Ability, Activities, and Questions for Verity BDD
sidebar:
  order: 10
---

A custom **Ability** extends Verity BDD to interact with any external interface your system uses —
a database, a message queue, a file system, a gRPC service, or anything else. Think of it as literally an ability of an actor to interact with a system or do specific actions in a system.

This guide walks through building a `FileSystemAbility` from scratch.
You'll end up with a fully functional ability including typed interactions and questions.

## What you're building

Actor has a method `WhoCan` which gives the actor ability to interact with a system. For example:
```go
// Give an actor the ability to manage files
actor := test.ActorCalled("Auditor").WhoCan(ManageFilesIn(t.TempDir()))

// Use typed interactions and questions
actor.AttemptsTo(
    WriteFile("report.txt", "audit passed"),
    ensure.That(FileContent("report.txt"), expectations.ContainsSubstring("audit")),
    ensure.That(FileExists("report.txt"), expectations.Equals(true)),
)
```

In this case `ManageFilesIn(t.TempDir())` returns a domain-specific ability interface. Let's build it.

## Step 1 — Define the interface

`verity.Ability` is currently an empty interface, so every Go type already satisfies it. Defining a narrower domain interface gives activities and questions a useful typed contract. Embedding `verity.Ability` is an optional marker convention:
Define the operations the ability exposes to interactions and questions:

```go title="filesystem/ability.go"
package filesystem

import (
    "fmt"
    "os"
    "path/filepath"
    "sync"

    verity "github.com/verity-bdd/verity-bdd"
)

// FileSystemAbility enables an actor to interact with the file system.
type FileSystemAbility interface {
    verity.Ability

    ReadFile(path string) (string, error)
    WriteFile(path string, content string) error
    DeleteFile(path string) error
    Exists(path string) bool
}
```

The operations are your domain contract; the embedded marker does not add methods or enforce conformance.

## Step 2 — Implement the ability

Create a private struct that holds the ability's state and implements the interface.
Use a mutex to make it safe for concurrent use:

```go title="filesystem/ability.go"
type fileSystemAbility struct {
    workingDir string
    mutex      sync.RWMutex
}

// ManageFilesIn creates a FileSystemAbility rooted at the given directory.
func ManageFilesIn(directory string) FileSystemAbility {
    if !filepath.IsAbs(directory) {
        if abs, err := filepath.Abs(directory); err == nil {
            directory = abs
        }
    }
    return &fileSystemAbility{workingDir: directory}
}

func (f *fileSystemAbility) ReadFile(path string) (string, error) {
    f.mutex.Lock()
    defer f.mutex.Unlock()

    data, err := os.ReadFile(filepath.Join(f.workingDir, path))
    if err != nil {
        return "", fmt.Errorf("failed to read file %s: %w", path, err)
    }
    return string(data), nil
}

func (f *fileSystemAbility) WriteFile(path string, content string) error {
    f.mutex.Lock()
    defer f.mutex.Unlock()

    fullPath := filepath.Join(f.workingDir, path)
    if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
        return fmt.Errorf("failed to create directory for %s: %w", path, err)
    }
    if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
        return fmt.Errorf("failed to write file %s: %w", path, err)
    }
    return nil
}

func (f *fileSystemAbility) DeleteFile(path string) error {
    f.mutex.Lock()
    defer f.mutex.Unlock()

    if err := os.Remove(filepath.Join(f.workingDir, path)); err != nil {
        return fmt.Errorf("failed to delete file %s: %w", path, err)
    }
    return nil
}

func (f *fileSystemAbility) Exists(path string) bool {
    f.mutex.RLock()
    defer f.mutex.RUnlock()

    _, err := os.Stat(filepath.Join(f.workingDir, path))
    return err == nil
}
```

## Step 3 — Add Activities

Activities are how actors use the ability inside `AttemptsTo`.
Each activity is a struct that implements three methods:

| Method | Returns | Purpose |
|---|---|---|
| `PerformAs(ctx, actor)` | `error` | The activity logic |
| `Description()` | `string` | Label shown in test reports |
| `FailureMode()` | `verity.FailureMode` | How to handle failures |

```go title="filesystem/activities.go"
package filesystem

import (
    "context"
    "fmt"

    verity "github.com/verity-bdd/verity-bdd"
)

// WriteFileActivity writes content to a file.
type WriteFileActivity struct {
    path    string
    content string
}

func WriteFile(path, content string) *WriteFileActivity {
    return &WriteFileActivity{path: path, content: content}
}

func (w *WriteFileActivity) PerformAs(ctx context.Context, actor verity.Actor) error {
    ability, err := actor.AbilityTo(&fileSystemAbility{})
    if err != nil {
        return fmt.Errorf("actor does not have FileSystemAbility: %w", err)
    }
    return ability.(FileSystemAbility).WriteFile(w.path, w.content)
}

func (w *WriteFileActivity) Description() string {
    return fmt.Sprintf("writes file: %s", w.path)
}

func (w *WriteFileActivity) FailureMode() verity.FailureMode {
    return verity.FailFast
}
```

The critical line is `actor.AbilityTo(&fileSystemAbility{})`. Ability lookup walks abilities in insertion order and returns the first assignable match. Matching supports concrete assignability, interface implementation, and corresponding pointer and value forms; pass a value representing the desired type, such as the zero-value pointer above.

`WhoCan` appends abilities and does not replace or deduplicate an existing type. This makes overlapping interface matches order-dependent. Verity stores the references it is given, so the caller or default-ability factory owns isolation: construct one mutable ability per actor unless sharing is intentional.

The simpler alternative is `verity.AbilityOf[FileSystemAbility](actor)`, used below. Request the interface itself, not `*FileSystemAbility` (a pointer to an interface).

Add activities for the other operations the same way:

```go title="filesystem/activities.go"
// DeleteFileActivity deletes a file.
type DeleteFileActivity struct {
    path string
}

func DeleteFile(path string) *DeleteFileActivity {
    return &DeleteFileActivity{path: path}
}

func (d *DeleteFileActivity) PerformAs(ctx context.Context, actor verity.Actor) error {
    ability, err := verity.AbilityOf[FileSystemAbility](actor) // returns FileSystemAbility, not verity.Ability
    if err != nil {
        return fmt.Errorf("actor does not have FileSystemAbility: %w", err)
    }
    // you don't have to cast it
    return ability.DeleteFile(d.path)
}

func (d *DeleteFileActivity) Description() string {
    return fmt.Sprintf("deletes file: %s", d.path)
}

func (d *DeleteFileActivity) FailureMode() verity.FailureMode {
    return verity.FailFast
}
```

## Step 4 — Add Questions

Questions let actors retrieve information from the ability and use it in assertions.
Each question implements `AnsweredBy` and `Description`:

```go title="filesystem/questions.go"
package filesystem

import (
    "context"
    "fmt"

    verity "github.com/verity-bdd/verity-bdd"
)

// FileContentQuestion retrieves the content of a file.
type FileContentQuestion struct {
    path string
}

func FileContent(path string) *FileContentQuestion {
    return &FileContentQuestion{path: path}
}

func (q *FileContentQuestion) AnsweredBy(ctx context.Context, actor verity.Actor) (string, error) {
    ability, err := actor.AbilityTo(&fileSystemAbility{})
    if err != nil {
        return "", fmt.Errorf("actor does not have FileSystemAbility: %w", err)
    }
    return ability.(FileSystemAbility).ReadFile(q.path)
}

func (q *FileContentQuestion) Description() string {
    return fmt.Sprintf("content of file: %s", q.path)
}
```

```go title="filesystem/questions.go"
// FileExistsQuestion checks whether a file exists.
type FileExistsQuestion struct {
    path string
}

func FileExists(path string) *FileExistsQuestion {
    return &FileExistsQuestion{path: path}
}

func (q *FileExistsQuestion) AnsweredBy(ctx context.Context, actor verity.Actor) (bool, error) {
    ability, err := actor.AbilityTo(&fileSystemAbility{})
    if err != nil {
        return false, fmt.Errorf("actor does not have FileSystemAbility: %w", err)
    }
    return ability.(FileSystemAbility).Exists(q.path), nil
}

func (q *FileExistsQuestion) Description() string {
    return fmt.Sprintf("existence of file: %s", q.path)
}
```

## Step 5 — Use it in a test

```go title="audit_test.go"
package audit_test

import (
    "context"
    "testing"

    verity "github.com/verity-bdd/verity-bdd"
    expectations "github.com/verity-bdd/verity-bdd/verity_expectations"
    "github.com/verity-bdd/verity-bdd/verity_expectations/ensure"

    "myproject/filesystem"
)

func TestAuditReport(t *testing.T) {
    ctx := context.Background()
    test := verity.NewVerityTestWithContext(ctx, t)

    auditor := test.ActorCalled("Auditor").WhoCan(
        filesystem.ManageFilesIn(t.TempDir()),
    )

    auditor.AttemptsTo(
        filesystem.WriteFile("report.txt", "audit passed"),
        ensure.That(filesystem.FileExists("report.txt"), expectations.Equals(true)),
        ensure.That(filesystem.FileContent("report.txt"), expectations.ContainsSubstring("audit")),
        filesystem.DeleteFile("report.txt"),
        ensure.That(filesystem.FileExists("report.txt"), expectations.Equals(false)),
    )
}
```

## Combining with other abilities

An actor can hold multiple abilities at once.
Each ability is retrieved independently by type, so they don't interfere:

```go
actor := test.ActorCalled("IntegrationTester").WhoCan(
    filesystem.ManageFilesIn(t.TempDir()),
    api.CallAnApiAt("https://api.example.org"),
)

// Fetch data from the API, then save it to disk
actor.AttemptsTo(api.SendGetRequest("/report"))

body, _ := api.LastResponseBody{}.AnsweredBy(ctx, actor)

actor.AttemptsTo(
    filesystem.WriteFile("response.json", body),
    ensure.That(filesystem.FileContent("response.json"), expectations.ContainsSubstring("status")),
)
```

## Summary

| Piece | What to implement |
|---|---|
| **Ability interface** | Declare the operations activities and questions need; embedding `verity.Ability` is optional |
| **Private struct** | Hold state, use `sync.RWMutex`, implement the interface |
| **Constructor** | Return the interface type, not the struct pointer |
| **Activity** | `PerformAs`, `Description`, `FailureMode` — retrieve via `actor.AbilityTo(&yourStruct{})` |
| **Question** | `AnsweredBy`, `Description` — same retrieval pattern |

For another complete, focused implementation using `QuestionAbout` and `AbilityOf`, see [Ability Examples](/en/examples/abilities/).
