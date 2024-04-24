# Flaming Zombies

A scheduler for performing monitoring checks, detecting failure and raising alerts.

---

THIS IS STILL A WORK IN PROGRESS

---


tasks -> gate -> notifier


Tasks are ran on a supplied schedule. The state of the task is supplied to one or more gates which determine whether the notifier should raise an alert. Tasks, Gates and Notifiers are all expected to be supplied by the user as short scripts (or complicated problems, if you prefer).


## Tasks

They define the test that is being performed.

A simple task definition might look like this:

```toml
[[task]]
name = "disk_percent_free_root"
command ="plugins/disk_free"
args = [ "/", "90" ]
frequency_seconds = 60
timeout_seconds = 5
```

This tasks does nothing but sleep for four seconds every minute. The name of the task is `sleep`. It's only used to make the logs a little easier to debug. The command that is run will be `sleep 4`.

This task a little pointless since it will never fail.



Tasks are scheduled. When they change state, the configured notifier is executed.

Both a task and a notifier is a stand-alone script or program which the user is expected to supply.


A simple example to alert if the disk is too full:

```bash
#!/usr/bin/env bash
[[ $(df -P | awk -v fs=$1'($6==fs) { print int($5) }') -gt $2 ]]
```

And you would configure a task to perform the check like so:

```toml
[[task]]
name = "disk_free:/"
command = "plugins/task/df"
args = [ "/", "90" ]
frequency_seconds = 3600
error_body = """
the disk at / is more than 90% full
"""
recover_body = """
the disk usage of / fallen below the threshold.
"""
```



## Configuration

The configuration is fetched from the supplied TOML file.


## Features



## Limitations

Tasks can run as frequently has every second.



---

The OpenBSD rc script can be found at: https://gist.github.com/jonhiggs/b3949cb5f7fd51997023c4c006eca2d5
