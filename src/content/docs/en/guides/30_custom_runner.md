---
title: Custom Test Runner
description: Uses any `*testing.T` wrapper
sidebar:
  order: 5
---

## Bring our own runner
There are frameworks out there, that wrap standard `*testing.T` type into a custom wrapper. You can use those wrappers to create a `VerityTest` as long as the resulting object satisfies the following interface:
```go
type TestContext interface {
	// Name returns the name of the test
	Name() string

	// Logf logs a formatted message
	Logf(format string, args ...interface{})

	// Errorf logs a formatted error message and marks the test as failed
	Errorf(format string, args ...interface{})

	// FailNow marks the test as failed and stops execution
	FailNow()

	// Failed returns true if the test has already failed
	Failed() bool

    // Specifies tear down logic
	Cleanup(func())

    // Lets the runner and the compiler know, that the method this
    // function is called in is a helper method
	Helper()
}
```
