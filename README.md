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

## Structure
Your calendars and their events are stored in a directory *root*, along with the calendar configuration.

The root defaults to `.ian` in your home directory (`~/.ian`).

The calendars are purely subdirectories inside the root. Creating a new calendar can be done by merely creating a directory (in the root).
Events are readable TOML files inside respective calendar.

This plain approach to calendaring makes calendar manipulation trivial, as it can be managed in a filesystem.
There are no hidden ties between the client and the event files: the events can be freely moved around and renamed.

An ian instance may be structured like this:

```
.ian
├── .config.toml
├── work
│   └── Harass co-workers
└── home
    ├── Something I'll get around to
    └── Be with Molgan
```

In the above example there are two calendars (`work` and `home`), with a total of three events.

Every calendar is configurable.

## Configuration
ian works without configuration. But it is probably for the configuration that you use ian, so you can call it essential.

There are two configuration files used in the client:
* Your preferences (defaults to `~/.ian.toml`).
* The calendar configuration (`.config.toml` inside ian root).

### Preference configuration
Your personal preferences (first day of the week, time zone, etc.) can be configured in your local preference file, which defaults to `~/.ian.toml`.
Here is an example configuration that lists all available preferences.

```toml
# ~/.ian.toml
root = "~/.ian" # ian directory that the client should use
timezone = "UTC" # defaults to your machine's time zone
no-validation = false # if true, disables event validation. helpful to effectively remedy corrupt events.
no-collision = false # if true, does not allow you to create events that chronologically collide with another event
collision-exceptions = ["birthdays"] # a list of calendars that will be ignored by the collision checker
no-collision-warnings = false # if true, warnings about colisions will not appear
first-weekday = 2 # first weekday of the week; 1 = Sunday, 2 = Monday, ... 7 = Saturday
weeks = true # if true, week numbers will be displayed where relevant in the calendar
months = 3 # amount of months to show at once in the calendar. to only show the current month, set to `1`.
no-timeline = false # if true, the timeline next to the calendar will be hidden
no-event-coloring = false # if true, the calendar day numbers (1-31) will not be colored according to the calendar's color of the events occurring that day
daywidth = 3 # the width each calendar day gets. the 'cal' UNIX command has a daywidth of 2.
no-legend = false # if true, the legend displaying the events' calendars and their colors is hidden
```

Any preferences here can also be overrridden per command with flags. For example, `weeks = true` in the configuration can be enabled temporarily with `ian --weeks`, or disabled temporarily with `ian --weeks=false`.

### Calendar configuration
The calendar configuration (`.config.toml` inside ian root), is more relevant to your workflow.

This configuration houses how your calendar works, like subscribed calendars, synchronization hooks, etc.
If you will be running an ian server, it is all configured in this file, meaning this file is shared between both client and server.

#### Sources
Sources are calendars that are not managed in your local instance. They are listed by name in `sources`.

A source can be a static iCalendar file (like a schedule, or a someone's shared calendar), or a CalDAV calendar (like someone's shared calendar that you can edit).
Each source is cached and updated. When a cached calendar has reached its `lifetime`, it will be downloaded anew.

```toml
[sources.joe]
  source = "https://calendar.example.com/share/3497503452398461/joes-calendar"
  type = "ical" # a static calendar
  lifetime = "47h30m" # optional: interval between cache updates (defaults to 2 hours)

# one more time!
[sources.mary]
  source = "webcal://canoga-park.net/caldav/mary"
  type = "caldav" # an editable calendar
```

| Attribute | Value             | Description                                   | Example                          | Required | Default |
|-----------|-------------------|-----------------------------------------------|----------------------------------|----------|---------|
| source    |iCal/WebCal URL    | URL to download cache from, or CalDAV server. |`https://example.com/schedule.ics`|          |         |
| type      |`ical` or `caldav` | Type of source.                               |`ical`                            |          |         |
| lifetime  |`_h_m_s` lifetime  | For how long the source should be cached.     |`3h40m`                           | optional | 2h      |

#### Calendars
In `calendars`, you can configure the behavior of both local and cached calendars (from sources).

```toml
[calendars.work]
  color = { r = 153, g = 90, b = 209 }
```

| Attribute | Value   | Description                    | Example                       | Required | Default |
|-----------|---------|--------------------------------|-------------------------------|----------|---------|
| color     |RGB color| Color for calendar recognition.| `{ r = 130, g = 49, b = 168 }`| optional | white   |

#### Hooks
Hooks are commands that perform wanted operations when the calendar is updated.

The main use for hooks is synchronizing your instance with a version control system (VCS), like [git](https://git-scm.com), but it can be used for anything that your shell can do.

Each hook is listed by name in `hooks` (e.g. `hooks.peter`).

| Attribute  | Value             | Description                                     | Example                                 | Required | Default |
|------------|-------------------|-------------------------------------------------|-----------------------------------------|----------|---------|
| precommand |Shell command      | Shell command executed before files are updated.|`echo "before: $(date) $MESSAGE" >> log` | optional |         |
| postcommand|Shell command      | Shell command executed after files are updated. |`echo "after:  $(date) $MESSAGE" >> log` | optional |         |
| type       |Bitmask (integer)  | What type of updates the hook should react on; 0 = any, 1 = ping (manual sync), 2 = event created, 4 = event updated, 8 = event deleted. Sum multiple to combine them.                                                  |`10` (only on creation and deletion)      |          |         |
| cooldown   |`_h_m_s` cooldown  | Time to wait before executing again.            |`1h`                                     | optional | `0s`    |

Keep in mind that manual file operations do not trigger these hooks. Only operations performed by the ian client or server do.

A manual sync to trigger these commands is possible with `ian sync`. If the `--ignore-cooldowns` (`-i`) flag is passed, all hooks will be triggered regardless of their cooldown status.

Cooldowns information is kept in the file `.cooldown-journal.toml`. Delete the file to reset the cooldowns.

##### Commands
The `precommand` command is executed BEFORE the changes are made, and `postcommand` AFTER.
Both commands are given a set of context environment variables:
* `MESSAGE`, a detailed message demonstrating what happened. Great for a git commit message.
* `FILES`, a list of space-separated files affected. If the type is a manual sync (1), this is empty.
* `TYPE`, the type of event that occured. This is not a bitmask, only one value (1, 2, 4, or 8).
The commands are executed inside the ian root directory.

## Usage
