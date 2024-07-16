![ian logo](assets/logo-small.svg "ian")
# ian

(gregor)ian is a UNIX-oriented, file-based, raw, calendar system CLIent and server.
ian was designed to quickly manipulate and manage personal and professional calendars.
A server allows you to centralize your events, instead of using your pocket calendar for that.

Events are stored in a (preferably version controlled) directory, with subdirectories for each calendar.
Each event is its own TOML file, which makes manual event manipulation easy.

This project strives for [iCalendar RFC 5545](https://datatracker.ietf.org/doc/html/rfc5545) compatability.

ian supports both interactive CalDAV calendars and simple read-only iCalendar calendars.

Synchronization is done via configuring command hooks.

## Setup
Clone the repository:
```
$ git clone https://github.com/truecrunchyfrog/ian
$ cd ian
```
Build the program:
```
$ go build -o ian github.com/truecrunchyfrog/ian/main
```
There should now be a binary in the working directory (`ian`, or `ian.exe` for Windows).

Next, to make the program available in your shell, either put the binary in an existing `$PATH` like:
```
$ mv ian /usr/bin
```
or add the source directory to your `$PATH`, by editing your `~/.profile`:
```
$ echo 'export PATH=$PATH:'$PWD >> ~/.profile
```
Now the program should be available in your shell. Try opening a new shell and running `ian`.
