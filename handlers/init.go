package handlers

import (
	"github.com/Sirupsen/logrus"
	"github.com/panjf2000/goproxy/config"
	"os"
	"github.com/panjf2000/goproxy/tool"
)

var conf config.Config

//this method will initialize a log module
func initLog() {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.TextFormatter{})

	// Output to stderr instead of stdout, could also be a file.
	logrus.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	logrus.SetLevel(logrus.DebugLevel)
	if !tool.CheckFileIsExist("logs") {
		os.Mkdir("logs", 0777)
	}

}

func init() {
	err := conf.GetConfig("config/config.json")
	initLog()
	if err != nil {
		logrus.Fatal(err)
	}
}
