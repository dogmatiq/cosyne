# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog], and this project adheres to
[Semantic Versioning].

<!-- references -->
[Keep a Changelog]: https://keepachangelog.com/en/1.0.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html
[BC]: https://github.com/dogmatiq/.github/blob/main/VERSIONING.md#changelogs

## [0.4.0] - 2026-11-12

### Added

- Add `Mutex`, a variant of `sync.Mutex` whose `Lock()` method stops waiting
  when its context is done. Unlike `golang.org/x/sync/semaphore`, it's ready to
  use as its zero value.

## [0.3.0] - 2026-11-05

This release replaces the package's contents. Its previous mutexes were
superseded by `sync.Mutex.TryLock()` and `golang.org/x/sync/semaphore`.

### Added

- Add `Signal`, which wakes a single waiting goroutine. A notification sent while
  nothing is waiting stays pending until the next wait, and repeated
  notifications coalesce into one.
- Add `SignalR` and `SignalW`, the read-only and write-only views of `Signal`.
- Add `Latch`, a one-way flag that unblocks all waiters when it is set.
- Add `LatchR` and `LatchW`, the read-only and write-only views of `Latch`.
- Add `Future`, a value that becomes available later, like a promise that cannot
  fail. It is stored at most once.
- Add `FutureR` and `FutureW`, the read-only and write-only views of `Future`.
- Add `Observable`, an atomic value that can be observed for changes.
- Add `ObservableR` and `ObservableW`, the read-only and write-only views of
  `Observable`.

### Changed

- **[BC]** The minimum supported Go version is now v1.27.

### Removed

- **[BC]** Remove `Mutex` and `RWMutex`.

### Fixed

- Fix `Once.Do()` choosing at random between calling the function and returning
  the context's error when the context is already done. It now returns the
  error without calling the function.

## [0.2.0] - 2021-01-02

### Added

- Add `Once`, a context-aware and "failable" variant of `sync.Once`
- Add `RWMutex.TryLock()` and `TryRLock()`

## [0.1.0] - 2020-05-20

- Initial release

<!-- references -->
[Unreleased]: https://github.com/dogmatiq/cosyne
[0.1.0]: https://github.com/dogmatiq/dogma/releases/tag/v0.1.0
[0.2.0]: https://github.com/dogmatiq/dogma/releases/tag/v0.2.0
[0.3.0]: https://github.com/dogmatiq/cosyne/releases/tag/v0.3.0
[0.4.0]: https://github.com/dogmatiq/cosyne/releases/tag/v0.4.0

<!-- version template
## [0.0.1] - YYYY-MM-DD

### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security
-->
