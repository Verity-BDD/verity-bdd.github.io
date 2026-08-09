---
title: Installation
description: Install Go and Verity BDD
sidebar:
  order: 1
---

This site documents **Verity BDD v0.22.3**, built from library commit [`5133ec7688f9f12d9ee581fe4311d1524dd2294f`](https://github.com/verity-bdd/verity-bdd/commit/5133ec7688f9f12d9ee581fe4311d1524dd2294f). The repository's `documented-library.json` manifest is the authoritative version declaration used by documentation checks and deployment.

:::note[v0.x.x notice]
  This project is still at version 0.x.x. No backwards compatibility is guaranteed for any changes. The plan is to go v1.x.x in Summer 2026.
:::

## Prerequisites

Verity BDD requires **Go 1.23.4 or later**.

If you don't have Go installed, follow the official installation guide at [go.dev/doc/install](https://go.dev/doc/install).

To verify your installation:

```bash
go version
# go version go1.23.4 darwin/arm64
```

## Create a new Go project

Create a directory for your project and initialize a Go module:

```bash
mkdir my-project
cd my-project
go mod init github.com/yourorg/my-project
```

Replace `github.com/yourorg/my-project` with your module path. For local-only projects, a simple name like `myproject` works fine.

Verity BDD tests live alongside your code and use the standard Go test file naming convention — any file ending in `_test.go`. You can place them anywhere in your module:

```
my-project/
├── go.mod
└── tests/
    └── create_post_test.go
```

## Install Verity BDD

In your Go module, run:

```bash
go get github.com/verity-bdd/verity-bdd@v0.22.3
```

## Package overview

| Package | Description |
|---|---|
| `github.com/verity-bdd/verity-bdd` | Core Screenplay API: actors, tasks, interactions, questions |
| `verity_abilities/api` | HTTP API ability, request activities, and response questions |
| `verity_abilities/take_notes` | Typed state shared between an actor's steps |
| `verity_abilities/wait` | Polling and channel-wait activities |
| `verity_answerable` | `ValueOf`, for wrapping a static value as a question |
| `verity_expectations` | Built-in expectations (`Equals`, `ContainsSubstring`, `ContainsKey`, `Satisfies`, …) |
| `verity_expectations/ensure` | `ensure.That` assertion activity |
| `verity_reporting` | Reporter, result, status, and attachment contracts |
| `verity_reporting/console_reporter` | Console reporter (included by default) |
| `verity_reporting/allure_reporter` | Allure reporter for CI reporting |

## Next steps

Once installed, head to [Writing Your First Test](/en/get_started/02_writing-your-first-test/) to write your first Screenplay-style API test.
