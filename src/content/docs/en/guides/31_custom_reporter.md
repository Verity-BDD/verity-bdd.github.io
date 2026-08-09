---
title: Custom reporter
description: Build your own test reporter
sidebar:
  order: 31
---

## Building a custom reporter

Any type that implements the `Reporter` interface can be passed to `Scene.Reporter`. The interface has five methods:

```go
type Reporter interface {
    OnTestStart(testName string)
    OnTestFinish(result TestResult)
    OnStepStart(stepDescription string)
    OnStepFinish(stepResult TestResult)
    SetOutput(w io.Writer)
}
```

`TestResult` provides the data each callback receives:

```go
type TestResult interface {
    Name()        string
    Status()      Status   // StatusPassed, StatusFailed, StatusSkipped
    Duration()    float64  // seconds
    Error()       error
    Attachments() []Attachment
}
```

Import both from `"github.com/verity-bdd/verity-bdd/verity_reporting"`.

The following tutorial builds a `MarkdownReporter` that writes a `.md` file for each test run.

### Step 1 — Define the struct

`path` is where the Markdown file will be written. `buf` accumulates lines as the test runs. `output` is an optional override for the write destination (used in tests and by `SetOutput`):

```go title="markdown_reporter/reporter.go"
package markdown_reporter

import (
    "fmt"
    "io"
    "os"
    "strings"

    verity_reporting "github.com/verity-bdd/verity-bdd/verity_reporting"
)

type MarkdownReporter struct {
    path   string
    buf    strings.Builder
    output io.Writer
}

func NewMarkdownReporter(path string) *MarkdownReporter {
    return &MarkdownReporter{path: path}
}
```

### Step 2 — `OnTestStart`

Called once when the test begins. Reset the buffer and write an H2 heading:

```go title="markdown_reporter/reporter.go"
func (r *MarkdownReporter) OnTestStart(testName string) {
    r.buf.Reset()
    fmt.Fprintf(&r.buf, "## %s\n\n", testName)
}
```

### Step 3 — `OnStepStart` and `OnStepFinish`

`OnStepStart` fires before a step runs and receives the description. `OnStepFinish` fires after and receives a `TestResult` with the outcome.

We write each step only after it finishes so we can include the status icon:

```go title="markdown_reporter/reporter.go"
func (r *MarkdownReporter) OnStepStart(_ string) {}

func (r *MarkdownReporter) OnStepFinish(result verity_reporting.TestResult) {
    switch result.Status() {
    case verity_reporting.StatusPassed:
        fmt.Fprintf(&r.buf, "- ✅ %s (%.2fs)\n", result.Name(), result.Duration())
    case verity_reporting.StatusFailed:
        fmt.Fprintf(&r.buf, "- ❌ %s (%.2fs)\n", result.Name(), result.Duration())
    default:
        fmt.Fprintf(&r.buf, "- ⏭️  %s\n", result.Name())
    }
}
```

### Step 4 — `OnTestFinish`

Called once after all steps complete. Write the overall result and flush to disk:

```go title="markdown_reporter/reporter.go"
func (r *MarkdownReporter) OnTestFinish(result verity_reporting.TestResult) {
    var statusLine string
    switch result.Status() {
    case verity_reporting.StatusPassed:
        statusLine = "✅ PASSED"
    case verity_reporting.StatusFailed:
        statusLine = "❌ FAILED"
    default:
        statusLine = "⏭️  SKIPPED"
    }

    fmt.Fprintf(&r.buf, "\n**Result:** %s | **Duration:** %.2fs\n", statusLine, result.Duration())

    if result.Error() != nil {
        fmt.Fprintf(&r.buf, "\n**Error:** %s\n", result.Error())
    }

    if len(result.Attachments()) > 0 {
        fmt.Fprintf(&r.buf, "\n> %d attachment(s) recorded — see the [Attachments guide](/en/guides/21_attachments/) for details.\n", len(result.Attachments()))
    }

    w := r.output
    if w == nil {
        f, err := os.Create(r.path)
        if err != nil {
            return
        }
        defer f.Close()
        w = f
    }
    fmt.Fprint(w, r.buf.String())
}
```

`result.Attachments()` returns `[]Attachment`, each with a `Name string`, `ContentType string`, and `Content []byte`. The reporter above acknowledges them and delegates to the [Attachments guide](/en/guides/21_attachments/).

### Step 5 — `SetOutput`

Satisfies the interface. Redirects the write destination (useful for tests):

```go title="markdown_reporter/reporter.go"
func (r *MarkdownReporter) SetOutput(w io.Writer) {
    r.output = w
}
```

### Step 6 — Use it in a test

```go title="order_test.go"
package order_test

import (
    "context"
    "testing"

    verity "github.com/verity-bdd/verity-bdd"
    "myproject/markdown_reporter"
)

func TestPlaceOrder(t *testing.T) {
    reporter := markdown_reporter.NewMarkdownReporter("test-report.md")
    test := verity.NewVerityTest(t, verity.Scene{
        Context:  context.Background(),
        Reporter: reporter,
    })
    defer test.Shutdown()

    customer := test.ActorCalled("Alice")
    customer.AttemptsTo(
        // ... your activities
    )
}
```

### Step 7 — Sample output

After the test runs, `test-report.md` contains:

```markdown title="test-report.md"
## TestPlaceOrder

- ✅ Alice opens the shop (0.01s)
- ✅ Alice adds item to cart (0.02s)
- ✅ Alice checks cart total equals 29.99 (0.00s)

**Result:** ✅ PASSED | **Duration:** 0.03s
```

---

## Attachments

Both result callbacks expose `result.Attachments()`, whose `Attachment` values carry a `Name string`, `ContentType string`, and `Content []byte`. A non-empty `take_notes` notebook becomes one test-level `"notes"` JSON attachment delivered to `OnTestFinish`. Standard `Actor.AttemptsTo` execution does not add step attachments, but callers that own a low-level `verity_reporting.ActivityTracker` can pass explicit attachments to `Finish`, which delivers them with `OnStepFinish`.

The `MarkdownReporter` tutorial above counts attachments and points to the dedicated guide. The built-in `ConsoleReporter` prints attachment content inline; `AllureReporter` persists each as a separate file alongside the JSON result.

See the [Attachments guide](/en/guides/21_attachments/) for the current limitations and rendering example.
