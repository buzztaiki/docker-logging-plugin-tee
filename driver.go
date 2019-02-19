package main

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/docker/docker/daemon/logger"
)

type driver struct {
}

func newDriver() *driver {
	return &driver{}
}

func (d *driver) startLogging(file string, info logger.Info) {
}

func (d *driver) stopLogging(file string) {
}

func (d *driver) readLogs(info logger.Info, config logger.ReadConfig) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewBuffer([]byte{}))
}
