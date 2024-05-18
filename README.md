# Flaming Zombies

A simple yet extendable, lightweight monitoring daemon.

* Made for the command line.
* Configured with flat files.
* Complete documentation in man pages.
* Few dependencies.
* Stateless¹.
* Easily extended, and customised.
* Liberal 2-clause BSD license.

<small>
1. State of course exists, but it isn't persisted between restarts.
</small>

---

THIS IS STILL A WORK IN PROGRESS... but it's ready to test.

---

Flaming Zombies ties together three components with three distinct responsibilities; `tasks`, `notifiers` and `gates`.

* [`tasks`](libexec/task) check whether a condition is true or false, like if a host responds to pings?
* [`notifiers`](libexec/notifier) raise alerts.
* [`gates`](libexec/gate) control when a `notifier` may execute.

```mermaid
sequenceDiagram
    task->>+notifier: new measurement
    notifier->>gate: are you open?
    gate->>+notifier: no
    task->>+notifier: new measurement
    notifier->>gate: are you open?
    gate->>+notifier: yes
    notifier->>Human: notification sent
```

You're expected to have many tasks. Each task can have one or more notifiers. Each notifier can have one or more gates.

And that is the basis of Flaming Zombies.

## Documentation

The complete documentation is available in the [man pages](./man). You can read them in your shell before they're installed using using the command:

```
curl https://raw.githubusercontent.com/jonhiggs/flamingzombies/main/man/man1/fz.1 | man /dev/stdin
```

## Building

Before you can build `fz` and `fzctl`, you'll need to have [Go](https://go.dev/doc/install) installed on your system.

To build, run:

```
go build ./cmd/fzctl/fzctl.go
go build ./cmd/fz/fz.go
```

That will produce the binaries for your system. The plugins are at `./libexec` and the man pages are at `./man`. Adapting the OpenBSD installation instructions should get you a long way to installing it on most UNIX-like system. You may find an init script for your operating system at `./scripts`. If you end up writing one, I would appreciate it if you could share it back.


## Installation

Installation is intended to be very simple. Eventually, I'd like to provide installation packages, but until then a manual process will need to suffice.

### OpenBSD

The below sequence of commands will install the daemon on OpenBSD:

```sh
## fz
wget https://github.com/jonhiggs/flamingzombies/releases/download/${VERSION}/fz_openbsd_${ARCH} \
    -O /usr/local/bin/fz

chown root:wheel /usr/local/bin/fz
chmod 755 /usr/local/bin/fz

## fzctl
wget https://github.com/jonhiggs/flamingzombies/releases/download/${VERSION}/fzctl_openbsd_${ARCH} \
    -O /usr/local/bin/fzctl

chown root:wheel /usr/local/bin/fzctl
chmod 755 /usr/local/bin/fzctl

## rc script
wget https://raw.githubusercontent.com/jonhiggs/flamingzombies/main/scripts/openbsd_rc \
    -O /etc/rc.d/flamingzombies

chown root:wheel /etc/rc.d/flamingzombies
chmod 755 /etc/rc.d/flamingzombies

## plugins
wget https://github.com/jonhiggs/flamingzombies/releases/download/${VERSION}/plugins.tar.gz \
    -O /tmp/plugins.tar.gz

tar -C /usr/local/libexec -zxvf /tmp/plugins.tar.gz
rm /tmp/plugins.tar.gz

## man pages
for m in man1/fz.1 man1/fzctl.1 man5/flamingzombies.toml.5 man7/fz-gates.7 man7/fz-notifiers.7 man7/fz-tasks.7; do
    wget https://raw.githubusercontent.com/jonhiggs/flamingzombies/main/man/$f \
        -O /usr/local/man/$f
done

## config
# create a configuration at /etc/flamingzombies.toml
# see the flamingzombies.toml(5) man page.

## enable the daemon
rcctl enable flamingzombies
rcctl set flamingzombies logger daemon.info
```

### Alpine Linux

The below sequence of commands will install the daemon on Alpine Linux:

```sh
## fz
wget https://github.com/jonhiggs/flamingzombies/releases/download/${VERSION}/fz_linux_${ARCH} \
    -O /usr/local/bin/fz

chown root:root /usr/local/bin/fz
chmod 755 /usr/local/bin/fz

## fzctl
wget https://github.com/jonhiggs/flamingzombies/releases/download/${VERSION}/fzctl_linux_${ARCH} \
    -O /usr/local/bin/fzctl

chown root:root /usr/local/bin/fzctl
chmod 755 /usr/local/bin/fzctl

## rc script
wget https://raw.githubusercontent.com/jonhiggs/flamingzombies/main/scripts/openrc \
    -O /etc/init.d/flamingzombies

chown root:root /etc/init.d/flamingzombies
chmod 755 /etc/init.d/flamingzombies

## plugins
wget https://github.com/jonhiggs/flamingzombies/releases/download/${VERSION}/plugins.tar.gz \
    -O /tmp/plugins.tar.gz

mkdir -p /usr/local/libexec
tar -C /usr/local/libexec -zxvf /tmp/plugins.tar.gz
rm /tmp/plugins.tar.gz

## man pages
for i in $(seq 1 7); do
    mkdir -p "/usr/local/man/man$i"
done

for m in man1/fz.1 man1/fzctl.1 man5/flamingzombies.toml.5 man7/fz-gates.7 man7/fz-notifiers.7 man7/fz-tasks.7; do
    wget https://raw.githubusercontent.com/jonhiggs/flamingzombies/main/man/$f \
        -O /usr/local/man/$f
done

## config
# create a configuration at /etc/flamingzombies.toml
# see the flamingzombies.toml(5) man page.

## enable the daemon
rc-update add flamingzombies
service flamingzombies start
```
