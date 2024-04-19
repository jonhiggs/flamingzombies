# plugins

A directory of task and notification plugins handlers.


## notifier

The interface for a notifier is:

- `SUBJECT`: The message subject.
- `PRIORITY`: The configured priority of the alert.
- `STATE`: The new state.
- `LAST_STATE`: the previous state.

- `stdin`: The notification body.