---
title: Attachments
description: Attach structured data to test steps and surface it in reports
sidebar:
  order: 21
---

Attachments are structured data the framework attaches to test results. Each attachment has three fields:

```go
type Attachment struct {
    Name        string
    ContentType string
    Content     []byte // serialised payload, e.g. JSON
}
```

The built-in high-level source is the [`take_notes` ability](/en/guides/11_notes/): when a test finishes, Verity BDD serialises all actor notes to JSON and delivers them as a single `"notes"` attachment in `OnTestFinish`. The standard actor pipeline has a narrower limitation: ordinary `Actor.AttemptsTo` activities call their tracker as `Finish(err)` and therefore cannot publish arbitrary step attachments.

The low-level reporting API is public. `verity_reporting.NewActivityTracker` and `NewActivityTrackerWithActor` return a tracker whose `Finish(err, attachments...)` accepts arbitrary `verity_reporting.Attachment` values. Use this API only when you own the execution/tracking boundary; do not wrap an activity already run through `Actor.AttemptsTo`, or you will report it twice.

```go
func reportCustomStep(reporter verity_reporting.Reporter, err error) {
    tracker := verity_reporting.NewActivityTracker(reporter, "exports diagnostics")
    tracker.Start()
    tracker.Finish(err, verity_reporting.Attachment{
        Name:        "diagnostics.json",
        ContentType: "application/json",
        Content:     []byte(`{"status":"captured"}`),
    })
}
```

## Receiving attachments in a reporter

Reporters can inspect `result.Attachments()` in either result callback. The high-level `take_notes` path adds its attachment to the test result passed to `OnTestFinish`; a low-level activity tracker can instead add explicit attachments to the step result passed to `OnStepFinish`:

```go
func (r *MyReporter) OnTestFinish(result verity_reporting.TestResult) {
    for _, att := range result.Attachments() {
        // att.Name        — e.g. "notes"
        // att.ContentType — e.g. "application/json"
        // att.Content     — raw bytes
    }
}
```

**When to expect attachments:**

| Callback | What arrives |
|---|---|
| `OnTestFinish` | Test-level attachments — includes `take_notes` output |
| `OnStepFinish` | Empty for standard `Actor.AttemptsTo`; low-level activity trackers can supply attachments explicitly |

**Processing rules:**
- `Content` is raw bytes. For `"application/json"`, `string(att.Content)` gives readable output.
- Use `ContentType` to decide how to render or persist.
- Always guard with `len(result.Attachments()) > 0` — tests without non-empty notes and standard `Actor.AttemptsTo` step results carry none, while low-level trackers may supply explicit step attachments.

## Example: rendering attachments in a Markdown reporter

The [custom reporter guide](/en/guides/31_custom_reporter/) builds a `MarkdownReporter` that counts attachments but doesn't render them. Here is the updated `OnTestFinish` that writes each attachment as a fenced block:

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
        fmt.Fprintf(&r.buf, "\n### Attachments\n\n")
        for _, att := range result.Attachments() {
            fmt.Fprintf(&r.buf, "**%s** (`%s`)\n\n```\n%s\n```\n\n",
                att.Name, att.ContentType, string(att.Content))
        }
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

When an actor uses `take_notes`, the generated `test-report.md` looks like:

```markdown title="test-report.md"
## TestPlaceOrder

- ✅ Alice opens the shop (0.01s)
- ✅ Alice adds item to cart (0.02s)
- ✅ Alice checks cart total equals 29.99 (0.00s)

**Result:** ✅ PASSED | **Duration:** 0.03s

### Attachments

**notes** (`application/json`)

\```
{"Alice":{"cart-total":"29.99","session-id":"abc123"}}
\```
```
