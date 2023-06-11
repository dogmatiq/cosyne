<div align="center">

# Cosyne

Cosyne provides [context](https://pkg.go.dev/context/?tab=doc)-aware
synchronization primitives for Go.

[![Documentation](https://img.shields.io/badge/go.dev-documentation-007d9c?&style=for-the-badge)](https://pkg.go.dev/github.com/dogmatiq/cosyne)
[![Latest Version](https://img.shields.io/github/tag/dogmatiq/cosyne.svg?&style=for-the-badge&label=semver)](https://github.com/dogmatiq/cosyne/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/dogmatiq/cosyne/ci.yml?style=for-the-badge&branch=main)](https://github.com/dogmatiq/cosyne/actions/workflows/ci.yml)
[![Code Coverage](https://img.shields.io/codecov/c/github/dogmatiq/cosyne/main.svg?style=for-the-badge)](https://codecov.io/github/dogmatiq/cosyne)

</div>

> **This project is deprecated.**
>
> This repository will be archived once it is no longer used by other Dogmatiq
> projects.

## Mutexes

Cosyne includes variants of `sync.Mutex` and `RWMutex` that accept a
`context.Context` parameter, which allows support for cancellation and deadlines
while acquiring locks.

The mutex implementations also support "try lock" semantics, allowing the user
to acquire the lock only if doing so would not block.
