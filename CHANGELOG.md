# CHANGELOG

## v0.0.21

- Support emitting output from tasks into notifications.
- Add `OPEN_GATES` to the environment of notifiers.
- Add `task/file_exists` plugin.
- Add `task/file_max_age` plugin.
- Add `task/loadavg` plugin.
- Add `task/swap_free` plugin.
- Add `task/tls_expiration` plugin.
- Modify the `task/disk_free` plugin to use kb, rather than percentage.
- Add tags to messages from `ntfy`.
- Improve timeout handling of task plugins.

## v0.0.20

- Many improvements to the docs and the interface.
- Removed dependency on logrus so that it can cross-compile for OpenBSD 7.5.
- Support gatesets to combine ANDing or ORing the results of gates.
- Improvements to the configtest
- Add `task/fz_errors` plugin
- Add `LAST_OK`, `LAST_FAIL` and `LAST_NOTIFICATION` to the environment of the gates.
- Support delayed notifications.
- Support re-raising notifications.
- Support detecting when a task is flapping.

## v0.0.19

- Add fzctl to check the daemon state
- Add scripts for testing HTTP
- Slight improvements to the man pages

## v0.0.18

- Release a tarball of plugins

## v0.0.17

- Test a release

## v0.0.16

- Simplify the release logic

## v0.0.15

- Improvements to the generated OpenBSD artifact
- Add a configtest mode.

## v0.0.14

- Remove dependency on goreleaser

## v0.0.13

- Fix crashes on OpenBSD

## v0.0.10

- License it under BSD-2-Clause
- Start a man page

## v0.0.7

- Add option parsing

## v0.0.5

- Another test release
