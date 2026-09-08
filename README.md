# backup-samurai

A backup orchestration tool. Largely inspired by [backupninja](https://0xacab.org/liberate/backupninja).

## Usage

```
./backup-samurai

  -v  --verbose           enable verbose logging
  -f  --force-execution   force execution, ignoring cron schedule
  -c  --config            path to config file
  -l  --log-file          path to logging file, use - to log to stdout (default)
      --no-update         disable update checker
```

## Cron schedule
This tool has a builtin cron-schedule-checker.
This is due to some devices (e.g. laptops) not being powered on when the cronjob would execute.
This way, the tool can additionally be ran at startup, and then internally check when it was last ran, running the jobs if the last execution was too long ago.

## Config

An example config can be found at `./config.toml`.
The individual tasks are executed as in order in the config file.

### Top-level
```toml
cron_schedule = ""    # Cron schedule which the tool should follow. Execution will be interrupted if tool is ran again before schedule hits.
```

### Handlers

#### restic

```toml
[handler.restic.name]
skip = false           # Whether to skip this step
ignore_fail = false    # if true, failing this step will not be logged as a failure
cron_schedule = ""     # custom cron schedule for this handler

repository = ""        # path to restic repository
password_file = ""     # path to password file
init = true            # if true, initialize the restic repo if needed
run_check = true       # if true, run `restic check`
run_prune = true       # if true, run `restic prune`
options = ""           # custom restic options
binary = ""            # custom restic binary

[handler.restic.name.backup]
run_backup = true      # if true, run `restic backup`
include = [""]         # files to include
exclude = [""]         # files to exclude

[handler.restic.name.forget]
run_forget = true      # if true, run `restic forget`
dry_run = false        # if true, dry run the forget process
keep_last = 0          # see restic forget documentation for more info (https://restic.readthedocs.io/en/latest/060_forget.html)
heep_hourly = 0
keep_daily = 0
keep_weekly = 0
keep_monthly = 0
keep_yearly = 0
```

#### pgbackup

```toml
[handler.pgbackup.name]
skip = false           # Whether to skip this step
ignore_fail = false    # if true, failing this step will not be logged as a failure
cron_schedule = ""     # custom cron schedule for this handler

database_user = ""     # database user
destination = ""       # output file or directory
gzip = true            # compress the output to single file
binary = ""            # custom pgbackup binary

[handler.pgbackup.name.docker]
run_in_container = true  # run the backup command inside docker container
container = ""         # container to run backup inside
```

#### rclone

```toml
[handler.rclone.name]
skip = false           # Whether to skip this step
ignore_fail = false    # if true, failing this step will not be logged as a failure
cron_schedule = ""     # custom cron schedule for this handler

binary = ""            # custom rclone binary

[handler.rclone.name.source]
path = ""              # location path
backend = ""           # rclone backend
flags = [""]           # custom cli flags

[handler.rclone.name.destination]
path = ""              # location path
backend = ""           # rclone backend
flags = [""]           # custom cli flags
```

#### rsync

```toml
[handler.rsync.name]
skip = false           # Whether to skip this step
ignore_fail = false    # if true, failing this step will not be logged as a failure
cron_schedule = ""     # custom cron schedule for this handler

binary = ""            # custom rclone binary

[handler.rsync.name.source]
remote = false         # true, if this is a remote location
location = ""          # location path
host = ""              # remote host
protocol = ""          # remote protocol
user = ""              # remote user

[handler.rsync.name.destination]
remote = false         # true, if this is a remote location
location = ""          # location path
host = ""              # remote host
protocol = ""          # remote protocol
user = ""              # remote user
```

#### script

```toml
[handler.script.name]
skip = false           # Whether to skip this step
ignore_fail = false    # if true, failing this step will not be logged as a failure
cron_schedule = ""     # custom cron schedule for this handler

binary = ""            # binary to run
options = [""]         # arguments to pass
```

### Notifiers

#### Notifiers & Update notifiers

All notifiers can also be used as update notifiers, by specifying them as `updatenotifier.type.name` instead of `notifier.type.name`.
Update notifiers will also be executed in order with other steps, so putting them at the very start or very end is recommended.

#### ntfy

Messages can be customized using data from the handlers, see ntfy.go for more info

```toml
[notifier.ntfy.name]
server = ""           # ntfy server to use
topic = ""            # ntfy topic to send to

[notifier.ntfy.name.authorization]
username = ""         # ntfy username
password = ""         # ntfy password
token = ""            # ntfy token, alternative to username and password

[notifier.ntfy.name.success]
title = ""            # message title
message = ""          # message content
priority = ""         # message priority
tags = ""             # message tags

[notifier.ntfy.name.failure]
title = ""            # message title
message = ""          # message content
priority = ""         # message priority
tags = ""             # message tags

[notifier.ntfy.name.update]
title = ""            # message title
message = ""          # message content
priority = ""         # message priority
tags = ""             # message tags
```

#### Webhook

```toml
[notifier.webhook.name]
url = ""              # URL to send to
method = ""           # HTTP method to use
meta = {""=""}        # Additional info to send
```