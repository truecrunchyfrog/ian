* replace arran4/golang-ical with emersion/go-ical
* implement CalDAV server
* implement CalDAV client
* migration: allow user to export/import calendars from/to ical, without adding as source. instead, use a command to migrate the ical file and provide a destination directory. also allow it the other way - exporting a native ian calendar to an ical. this will be more standalone than using sources, with more freedom of choice. this will work in tandem with archiving.
* path to proper path, not filepath
* togglable 24/12-hour clock
* create benchmarks and tests
* archiving
* create busy/free system and replace the "collision" system with it
* make showing an individual day's events easier (and a shorthand for today 'ian today' / 'ian now')
* make source usage safer; dont overwrite a calendar before you have it, and handle unavailable calendars better (allow them, but warn).
