# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog], and this project adheres to
[Semantic Versioning].

<!-- references -->
[Keep a Changelog]: https://keepachangelog.com/en/1.0.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html
[BC]: https://github.com/dogmatiq/.github/blob/main/VERSIONING.md#changelogs

## [Unreleased]

This release replaces the package's contents. Its previous mutexes were
superseded by `sync.Mutex.TryLock()` and `golang.org/x/sync/semaphore`.

### Added

- Add `Signal`, which notifies a waiting goroutine. Repeated notifications
  coalesce.
- Add `Latch`, a one-way flag that unblocks all waiters when it is set.
- Add `Future`, a value that is stored at most once and may not be available yet.
- Add `Observable`, an atomic value that can be observed for changes.
- Add `SignalR`, `LatchR`, `FutureR` and `ObservableR`, the read-only views of
  those types, and `SignalW`, `LatchW`, `FutureW` and `ObservableW`, the
  write-only views.

### Changed

- **[BC]** `Once.Do()` now returns the context's error, without calling the
  function, if the context is already done and the function has not succeeded.
- **[BC]** The minimum supported Go version is now v1.27.

### Removed

- **[BC]** Remove `Mutex` and `RWMutex`.

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

<!-- version template
## [0.0.1] - YYYY-MM-DD

### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security
-->
