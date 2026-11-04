<div align="center">

# Cosyne

Cosyne provides [context](https://pkg.go.dev/context/?tab=doc)-aware
synchronization primitives for Go.

[![Documentation](https://img.shields.io/badge/go.dev-documentation-007d9c?&style=for-the-badge)](https://pkg.go.dev/github.com/dogmatiq/cosyne)
[![Latest Version](https://img.shields.io/github/tag/dogmatiq/cosyne.svg?&style=for-the-badge&label=semver)](https://github.com/dogmatiq/cosyne/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/dogmatiq/cosyne/ci.yml?style=for-the-badge&branch=main)](https://github.com/dogmatiq/cosyne/actions/workflows/ci.yml)
[![Code Coverage](https://img.shields.io/codecov/c/github/dogmatiq/cosyne/main.svg?style=for-the-badge)](https://codecov.io/github/dogmatiq/cosyne)

</div>

Each primitive can be waited for with a `context.Context`, or used in a `select`
statement.

| Type            | Purpose                                                                    |
| --------------- | -------------------------------------------------------------------------- |
| `Signal`        | Notifies a waiting goroutine; repeated notifications coalesce.             |
| `Latch`         | A one-way flag that unblocks all waiters when set.                         |
| `Future[T]`     | A value that is stored at most once, and may not be available yet.         |
| `Observable[T]` | An atomic value that readers can observe for changes.                      |
| `Once`          | A variant of `sync.Once` that can fail, and can stop waiting for its turn. |

`Signal`, `Latch`, `Future` and `Observable` also have read-only and write-only
views, such as `LatchR` and `LatchW`, for handing out just one side of a value.
