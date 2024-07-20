* configuration
    * colorless mode
    * let user specify prefered datetime layouts
    * togglable 24/12-hour clock
* commands
    * default calendar
    * make the 'add' command friendlier; allow aliases (now, today, tomorrow, wednesday)
    * make it easier to specify what calendar to add events to
    * make rrule, rdate and exdate friendlier
    * 'name' flag for 'add' command
    * make showing an individual day's events easier (and a shorthand for today 'ian today' / 'ian now')
* ui
    * better calendar view. improve the side by side view (make a function for it).
    * select calendars/events to view
    * accentuate today and tomorrow/week in calendar view
    * create a universal event view, that will be shown when creating it, editing it, deleting it and showing info about it
    * list calendars command
* implementations
    * CalDAV
        * server
        * client
    * vCard (?)
    * vTodo
    * Zerolog (?)
* migration
    * allow user to export/import calendars from/to ical, without adding as source. instead, use a command to migrate the ical file and provide a destination directory. also allow it the other way - exporting a native ian calendar to an ical. this will be more standalone than using sources, with more freedom of choice. this will work in tandem with archiving.
    * archiving
    * create busy/free system and replace the "collision" system with it. check out event tranparency (TRANSP).
    * event categories
    * event statuses
* cleanup
    * public flag vars into cmd lookups; it's getting cluttery!
* cache
    * make sources work offline
    * make source usage safer; dont overwrite a calendar before you have it, and handle unavailable calendars better (allow them, but warn).
* create benchmarks and tests
