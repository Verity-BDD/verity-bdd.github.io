---
title: Reporting
description: Configure console and Allure reporting in Verity BDD, and build custom reporters
sidebar:
  order: 20
---

Verity BDD ships with two reporters ready to use. You can also build your own by implementing a simple five-method interface.

## ConsoleReporter

`ConsoleReporter` is the default reporter. It writes emoji-annotated, indented output to `stdout` as your tests run.

```go title="my_test.go"
import (
    verity "github.com/nchursin/verity-bdd"
    "github.com/nchursin/verity-bdd/verity_reporting/console_reporter"
)

func TestMyFeature(t *testing.T) {
    reporter := console_reporter.NewConsoleReporter()
    test := verity.NewVerityTest(t, verity.Scene{Reporter: reporter})
    // ...
}
```

Output:

```
🚀 Starting: TestMyFeature
  ✅ Sam sends GET /posts (0.12s)
  ✅ Sam checks response status equals 200 (0.00s)
✅ TestMyFeature: PASSED (0.13s)
```

### Writing to a file

Pass any `io.Writer` to `SetOutput`:

```go title="my_test.go"
f, _ := os.Create("report.txt")
defer f.Close()

reporter := console_reporter.NewConsoleReporter()
reporter.SetOutput(f)
```

## AllureReporter

`AllureReporter` writes [Allure 2](https://docs.qameta.io/allure/) result files that CI systems and the Allure dashboard can read.

```go title="my_test.go"
import (
    verity "github.com/nchursin/verity-bdd"
    "github.com/nchursin/verity-bdd/verity_reporting/allure_reporter"
)

func TestMyFeature(t *testing.T) {
    reporter := allure_reporter.NewAllureReporterWithDir("allure-results")
    test := verity.NewVerityTest(t, verity.Scene{Reporter: reporter})
    // ...
}
```

Each test produces a `{uuid}-result.json` file in the directory. Point your Allure CLI or dashboard at `allure-results/` to generate a visual report.

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

Import both from `"github.com/nchursin/verity-bdd/verity_reporting"`.

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

    verity_reporting "github.com/nchursin/verity-bdd/verity_reporting"
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

    verity "github.com/nchursin/verity-bdd"
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

Reporters receive attachments via `result.Attachments()` in `OnTestFinish` and `OnStepFinish`. Each `Attachment` carries a `Name string`, `ContentType string`, and `Content []byte` (a serialised payload such as JSON).

The `MarkdownReporter` tutorial above counts attachments and points to the dedicated guide. The built-in `ConsoleReporter` prints attachment content inline; `AllureReporter` persists each as a separate file alongside the JSON result.

For how activities *produce* attachments during execution, see the [Attachments guide](/en/guides/21_attachments/).
