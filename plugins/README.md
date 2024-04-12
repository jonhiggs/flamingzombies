# plugins

A directory of task and notification plugins handlers.


## notifier

The interface for a notifier is:

- `SUBJECT`: The message subject.
- `NAME`: The configured name of the notifier.
- `PRIORITY`: The configured priority of the alert.

- `stdin`: The notification body.
