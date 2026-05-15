---
title: Reporting
description: Configure console and Allure reporting in Verity BDD
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
