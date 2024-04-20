# libexec

A directory of scripts for handling tasks, notifications and gates.

## notifier

The interface for a notifier is:

- `SUBJECT`: The message subject.
- `PRIORITY`: The configured priority of the alert.
- `STATE`: The new state.
- `LAST_STATE`: the previous state.

- `stdin`: The notification body.
