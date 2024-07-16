![ian logo](assets/logo-small.svg "ian")
# ian

(gregor)ian is a UNIX-oriented, file-based, raw, calendar system CLIent and server.
ian was designed to quickly manipulate and manage personal and professional calendars.
A server allows you to centralize your events, instead of using your pocket calendar for that.

Events are stored in a (probably version controlled) directory, with subdirectories for each calendar.
Each event is its own TOML file, which makes manual event manipulation easy.

This project strives for [iCalendar RFC 5545](https://datatracker.ietf.org/doc/html/rfc5545) compatability.

ian supports dynamic and subscribeable calendars:
* Native ian calendars.
* Dynamic (interactive) CalDAV calendars.
* Static iCalendars.

Synchronization is done via configured listener commands.
