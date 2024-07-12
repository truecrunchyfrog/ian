* make instance.Events private and instead make a function to get events within a period, to support endless rrules
* change instance.ReadEvents parameter to apply to all events, not just recurring. and allow infinite range (both From and To are IsZero()), and when a recurring event is not satisfied (has more events to come) there should be a subtle note about that, however the client chooses to display it (e.g. at bottom of timeline)
* migration: allow user to export/import calendars from/to ical, without adding as source. instead, use a command to migrate the ical file and provide a destination directory. also allow it the other way - exporting a native ian calendar to an ical. this will be more standalone than using sources, with more freedom of choice. this will work in tandem with archiving.
* path to proper path, not filepath
* togglable 24/12-hour clock
* iCal Rrule support
* CalDAV support
* create benchmarks and tests
* archiving
* create busy/free system and replace the "collision" system with it
* make showing an individual day's events easier (and a shorthand for today 'ian today' / 'ian now')
