package handlers

import (
	log "github.com/Sirupsen/logrus"
	"github.com/panjf2000/goproxy/config"
	"os"
)

var conf config.Config

//this method will initialize a log module
func initLog() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stderr instead of stdout, could also be a file.
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)

}

func init() {
	err := conf.GetConfig("conf/config.json")
	initLog()
	if err != nil {
		log.Fatal(err)
	}
}
