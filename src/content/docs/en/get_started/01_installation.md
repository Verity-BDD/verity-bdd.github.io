---
title: Installation
description: Install Go and Verity BDD
sidebar:
  order: 1
---

:::note[v0.x.x notice]
  This project is still at version 0.x.x. No backwards compatibility is guaranteed for any changes. The plan is to go v1.x.x in Summer 2026.
:::

## Prerequisites

Verity BDD requires **Go 1.21 or later**.

If you don't have Go installed, follow the official installation guide at [go.dev/doc/install](https://go.dev/doc/install).

To verify your installation:

```bash
go version
# go version go1.21.0 darwin/arm64
```

## Install Verity BDD

In your Go module, run:

```bash
go get github.com/nchursin/verity-bdd
```

## Package overview

| Package | Description |
|---|---|
| `github.com/nchursin/verity-bdd` | Core Screenplay API: actors, tasks, interactions, questions |
| `verity_abilities/api` | HTTP API ability and built-in interactions |
| `verity_answerable` | Helpers for wrapping static and dynamic values as questions |
| `verity_expectations` | Built-in expectations (`Equals`, `Contains`, `Satisfies`, …) |
| `verity_expectations/ensure` | `ensure.That` assertion activity |
| `verity_reporting/console_reporter` | Console reporter (included by default) |
| `verity_reporting/allure_reporter` | Allure reporter for CI reporting |

## Next steps

Once installed, head to [Writing Your First Test](/en/guides/2_writing-your-first-test/) to write your first Screenplay-style API test.
