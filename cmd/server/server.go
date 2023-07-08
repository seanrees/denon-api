package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/seanrees/denon-api/pkg/avr"
)

var (
	host       = flag.String("host", "localhost", "host/IP to listen on")
	port       = flag.Int("port", 8080, "port to listen on")
	avrAddress = flag.String("avr", "", "host of avr to connect to")
	avrPort    = flag.Int("avr_port", 23, "port of avr")
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stderr)

	flag.Parse()

	r := gin.Default()

	addr := fmt.Sprintf("%s:%d", *avrAddress, *avrPort)
	denon := avr.NewAvr1912(addr)
	defer denon.Close()

	r.StaticFile("/", "assets/controller.html")

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// Input commands:
	r.GET("/input", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"input": denon.InputSource()})
	})

	r.PUT("/input/:input", func(c *gin.Context) {
		input := c.Param("input")
		input = denon.SetInputSource(strings.ToUpper(input))

		c.JSON(http.StatusOK, gin.H{"input": input})
	})

	// Surround Mode commands:
	r.GET("/mode", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"mode": denon.SurroundMode()})
	})

	r.PUT("/mode/:mode", func(c *gin.Context) {
		mode := c.Param("mode")
		mode = denon.SetSurroundMode(strings.ToUpper(mode))

		c.JSON(http.StatusOK, gin.H{"mode": mode})
	})

	// Power commands:
	r.GET("/power", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"power": denon.Power()})
	})

	r.PUT("/power/standby", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"power": denon.Standby()})
	})

	r.PUT("/power/on", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"power": denon.PowerOn()})
	})

	// Volume commands:
	r.GET("/volume", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"volume": denon.Volume()})
	})

	r.POST("/volume/up", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"volume": denon.VolumeUp()})
	})

	r.POST("/volume/down", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"volume": denon.VolumeDown()})
	})

	r.PUT("/volume/:level", func(c *gin.Context) {
		level := c.Param("level")
		v := denon.SetVolume(level)

		c.JSON(http.StatusOK, gin.H{"volume": v})
	})

	a := fmt.Sprintf("%s:%d", *host, *port)
	r.Run(a)
}
