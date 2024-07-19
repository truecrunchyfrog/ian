![ian logo](assets/logo-small.svg "ian")
# ian

(gregor)ian is a UNIX-oriented and file-based calendar system CLIent and server.
Its purpose is versatile and organized calendar manipulation and management.
A server is a replacement for proprietary and opaque SaaS applications (e.g. Google Calendar).

Events are stored in a (preferably version controlled) directory, with subdirectories for each calendar.
Each event is its own TOML file, which makes manual event manipulation easy.

This project strives for [iCalendar RFC 5545](https://datatracker.ietf.org/doc/html/rfc5545) compatability.

Both CalDAV and static iCalendar calendars are supported.

Synchronization is done via configuring command hooks.

## Installation
Clone the repository:
```
$ git clone https://github.com/truecrunchyfrog/ian
$ cd ian
```
Build the program:
```
$ go build -o ian github.com/truecrunchyfrog/ian/main
```
There should now be a binary in the working directory (`ian`).

For a single-user installation, you can add the source directory to your `$PATH`, by editing your `~/.profile` (or, e.g. `~/.bashrc`):
```
$ echo 'export PATH=$PATH:'$PWD >> ~/.profile
```
For a multi-user installation, you can place the binary in a system-wide `$PATH` directory like `/usr/bin`.

Now the program should be available in your shell. Try opening a new shell and running `ian`.

## Usage
ian works without configuration, but the configuration is probably why you use ian, so you can call it essential.
There are two configuration files used in the client:
* Your preferences (defaults to `~/.ian.toml`).
* The calendar configuration (defaults to `.ian/.config.toml`).

### Preference configuration
Your personal preferences (first day of the week, time zone, etc.) can be configured in your local preference file, which defaults to `~/.ian.toml`.
The options stored there are also available as flags, so you can override them per-command too.
```toml
# ~/.ian.toml
weeks = true
```
This file does normally not need to be updated often.

### Calendar configuration
The calendar configuration, which defaults to `.ian/.config.toml`, is a lot more important.

This configuration houses how your calendar works, like subscribed calendars, synchronization hooks, etc.

If you will be running an ian server, it is all configured in this file, meaning this file is shared between both client and server.
