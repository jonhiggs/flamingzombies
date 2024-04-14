# Flaming Zombies

A lightweight monitoring tool for small environments. It is more a monitoring scheduler than anything. It manages two concepts; tasks and notifiers.


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


