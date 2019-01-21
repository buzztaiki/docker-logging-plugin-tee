package main

import (
	"github.com/Sirupsen/logrus"
)

const pluginName = "tee"

var log *logrus.Entry

func init() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	log = logger.WithField("pluginName", pluginName)
}

func main() {
	h := newHandler()
	if err := h.ServeUnix(pluginName, 0); err != nil {
		panic(err)
	}
}
