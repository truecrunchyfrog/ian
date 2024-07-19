package server

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
	"github.com/gin-gonic/gin"
	"github.com/truecrunchyfrog/ian"
)

func Run(addrNative, addrCalDav string, debug bool, instance *ian.Instance) {
	go server(addrNative, debug, instance)
	go serverCalDav(addrCalDav, debug, instance)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println()
	log.Println("stopped")
}

func server(addr string, debug bool, instance *ian.Instance) {
	logger := log.New(os.Stderr, "[native server] ", 0)

	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	logger.Printf("starting native ian server on %s\n", addr)
	r.Run(addr)
}

type CalDavBackend struct {
	instance *ian.Instance
	logger   *log.Logger
}

func (backend CalDavBackend) CalendarHomeSetPath(ctx context.Context) (string, error) {
	return "", nil
}
func (backend CalDavBackend) CurrentUserPrincipal(ctx context.Context) (string, error) {
	return "default", nil
}
func (backend CalDavBackend) DeleteCalendarObject(ctx context.Context, path string) error {
	backend.logger.Printf("tried to delete object %s\n", path)
	return nil
}
func (backend CalDavBackend) GetCalendar(ctx context.Context, path string) (*caldav.Calendar, error) {
	backend.logger.Printf("tried to get calendar %s\n", path)
	return nil, nil
}
func (backend CalDavBackend) GetCalendarObject(ctx context.Context, path string, req *caldav.CalendarCompRequest) (*caldav.CalendarObject, error) {
	cal := ian.SanitizePath(path)

	events, _, err := backend.instance.ReadEvents(ian.TimeRange{})
	if err != nil {
		return nil, err
	}

	events = ian.FilterEvents(&events, func(e *ian.Event) bool {
		return e.Path.Calendar() == cal
	})

	var lastModified time.Time

	for _, event := range events {
		if mod := event.Props.Modified; mod.After(lastModified) {
			lastModified = mod
		}
	}

	var calName string
	if cal != "" {
		calName = cal
	} else {
		calName = "main"
	}

	ics := ian.ToIcal(events, calName)
	b, err := ian.SerializeIcal(ics)
	if err != nil {
		return nil, err
	}

	length := b.Len()

	h := sha1.New()
	if _, err := b.WriteTo(h); err != nil {
		return nil, err
	}
	csum := h.Sum(nil)

	calObj := caldav.CalendarObject{
		Path:          path,
		ModTime:       lastModified,
		ContentLength: int64(length),
		ETag:          base64.StdEncoding.EncodeToString(csum[:]),
		Data:          ics,
	}

	return &calObj, nil
}
func (backend CalDavBackend) ListCalendarObjects(ctx context.Context, path string, req *caldav.CalendarCompRequest) ([]caldav.CalendarObject, error) {
	backend.logger.Printf("tried to list objects in %s with %v\n", path, *req)
	return nil, nil
}
func (backend CalDavBackend) ListCalendars(ctx context.Context) ([]caldav.Calendar, error) {
	backend.logger.Println("tried to list calendars")
	return []caldav.Calendar{}, nil
}
func (backend CalDavBackend) PutCalendarObject(ctx context.Context, path string, calendar *ical.Calendar, opts *caldav.PutCalendarObjectOptions) (loc string, err error) {
	// grabbedAt is used to determine the age of the calendar the client has.
	// Only calendar events where the modified date is not after this time may be modified.
	grabbedAt, err := calendar.Props.DateTime(ian.IcalPropGrabTimestamp, time.UTC)
	if err != nil {
		return "", fmt.Errorf("missing ian property '%s' for revision\n", ian.IcalPropGrabTimestamp)
	}

	cal := ian.SanitizePath(path)

	// events are the current events.
	events, _, err := backend.instance.ReadEvents(ian.TimeRange{})
	events = ian.FilterEvents(&events, func(e *ian.Event) bool {
		return e.Path.Calendar() == cal
	})

	proposedEvents := calendar.Events()

	var hasPut bool

	for _, event := range events {
		i := slices.IndexFunc(proposedEvents, func(evProps ical.Event) bool {
			return evProps.Props.Get(ical.PropUID).Value == event.Props.Uid
		})

		if i == -1 {
			// No event match: delete.
			if event.Props.Modified.After(grabbedAt) {
				return "", errors.New("client wants to delete outdated event. synchronize changes first.")
			}

			file := event.Path.Filepath(backend.instance)

			err := backend.instance.Sync(func() error {
				return os.Remove(file)
			}, ian.SyncEvent{
				Type:    ian.SyncEventDelete,
				Files:   []string{file},
        Message: fmt.Sprintf("ian: [CalDAV request] delete event: '%s'", event.Path),
			}, false, nil)

			if err != nil {
				return "", err
			}
			hasPut = true
		}

		proposedEvent := proposedEvents[i]
		modified, err := proposedEvent.Props.DateTime(ical.PropLastModified, time.UTC)
		if err != nil {
			return "", err
		}
		if !modified.Equal(event.Props.Modified) {
			// Event match but properties changed: update.
			if event.Props.Modified.After(grabbedAt) {
				return "", errors.New("client wants to update outdated event. synchronize changes first.")
			}

			event.Props, err = ian.FromIcalEvent(proposedEvent)
			if err != nil {
				return "", err
			}

      err := backend.instance.Sync(func() error {
        return event.Write(backend.instance)
      }, ian.SyncEvent{
      	Type:    ian.SyncEventUpdate,
      	Files:   []string{event.Path.Filepath(backend.instance)},
        Message: fmt.Sprintf("ian: [CalDAV request] edit event '%s'", event.Path),
      }, false, nil)

      if err != nil {
        return "", err
      }
			hasPut = true
		}
	}

	if !hasPut {
		for _, proposedEvent := range proposedEvents {
			i := slices.IndexFunc(events, func(evProps ian.Event) bool {
				return evProps.Props.Uid == proposedEvent.Props.Get(ical.PropUID).Value
			})

			if i == -1 {
				// Event does not exist; create it.
				created, err := proposedEvent.Props.DateTime(ical.PropCreated, time.UTC)
				if err != nil {
					return "", err
				}
				if created.Before(grabbedAt) {
					return "", errors.New("event created before grab date. the event may have existed and was deleted, but not yet deleted for the client.")
				}
				props, err := ian.FromIcalEvent(proposedEvent)
				if err != nil {
					return "", err
				}
				event, err := backend.instance.WriteNewEvent(props, cal)
				if err != nil {
					return "", err
				}
				event.Write(backend.instance)
				hasPut = true
			}
		}
	}

	if !hasPut {
		return "", errors.New("no new/modified/deleted event")
	}

	return "", nil
}
func (backend CalDavBackend) QueryCalendarObjects(ctx context.Context, query *caldav.CalendarQuery) ([]caldav.CalendarObject, error) {
	backend.logger.Printf("tried to query objects with %v\n", *query)
	return nil, nil
}

func serverCalDav(addr string, debug bool, instance *ian.Instance) {
	logger := log.New(os.Stderr, "[CalDAV server] ", 0)
	logger.Printf("starting CalDAV server on %s\n", addr)

	var backend caldav.Backend = CalDavBackend{
		instance: instance,
		logger:   logger,
	}

	h := caldav.Handler{
		Backend: backend,
		Prefix:  "",
	}
	http.Handle("/", &h)
	logger.Fatal(http.ListenAndServe(addr, nil))
}
