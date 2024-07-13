package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	log.Printf("starting native ian server on %s\n", addr)
	r.Run(addr)
}

func serverCalDav(addr string, debug bool, instance *ian.Instance) {
	log.Printf("starting CalDAV server on %s\n", addr)

  panic("TODO: implement")
}
