package server

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
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
	log.Printf("tried to delete object %s\n", path)
	return nil
}
func (backend CalDavBackend) GetCalendar(ctx context.Context, path string) (*caldav.Calendar, error) {
	log.Printf("tried to get calendar %s\n", path)
	return nil, nil
}
func (backend CalDavBackend) GetCalendarObject(ctx context.Context, path string, req *caldav.CalendarCompRequest) (*caldav.CalendarObject, error) {
  cal := ian.SanitizePath(path)

	log.Printf("tried to read calendar %s with %v\n", cal, *req)

	events, _, err := backend.instance.ReadEvents(ian.TimeRange{})
	if err != nil {
		return nil, err
	}

  events = ian.FilterEvents(&events, func(e *ian.Event) bool {
    return e.GetCalendarName() == cal
  })

  var lastModified time.Time

  for _, event := range events {
    if mod := event.Props.Modified; mod.After(lastModified) {
      lastModified = mod
    }
  }

	ics := ian.ToIcal(events)
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
	log.Printf("tried to list objects in %s with %v\n", path, *req)
	return nil, nil
}
func (backend CalDavBackend) ListCalendars(ctx context.Context) ([]caldav.Calendar, error) {
	log.Println("tried to list calendars")
	return []caldav.Calendar{}, nil
}
func (backend CalDavBackend) PutCalendarObject(ctx context.Context, path string, calendar *ical.Calendar, opts *caldav.PutCalendarObjectOptions) (loc string, err error) {
	log.Printf("tried to put object %s in calendar %v with %v\n", path, *calendar, *opts)
	return "", nil
}
func (backend CalDavBackend) QueryCalendarObjects(ctx context.Context, query *caldav.CalendarQuery) ([]caldav.CalendarObject, error) {
	log.Printf("tried to query objects with %v\n", *query)
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
