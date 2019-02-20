package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/containerd/fifo"
	"github.com/docker/docker/daemon/logger"
	"github.com/docker/docker/daemon/logger/jsonfilelog"
	"github.com/pkg/errors"
)

type driver struct {
	streams map[string]io.ReadCloser
	loggers map[string]logger.Logger
	mu      sync.Mutex
}

func newDriver() *driver {
	return &driver{
		streams: map[string]io.ReadCloser{},
		loggers: map[string]logger.Logger{},
	}
}

func (d *driver) startLogging(file string, info logger.Info) error {
	if info.LogPath == "" {
		logDir := "/var/log/docker"
		if err := os.MkdirAll(logDir, 0700); err != nil {
			return errors.Wrap(err, "failed to create log dir")
		}
		info.LogPath = filepath.Join(logDir, info.ContainerID)
	}

	l, err := newLogger(info)
	if err != nil {
		return errors.Wrap(err, "failed to create logger")
	}
	r, err := openFifo(file)
	if err != nil {
		return errors.Wrap(err, "failed to open fifo")
	}

	d.mu.Lock()
	d.streams[file] = r
	d.loggers[info.ContainerID] = l
	d.mu.Unlock()

	go copyToLogger(r, l)
	return nil
}

func (d *driver) stopLogging(file string) error {
	d.mu.Lock()
	r, ok := d.streams[file]
	d.mu.Unlock()

	if !ok {
		return fmt.Errorf("stream for %s not found", file)
	}

	return r.Close()
}

func (d *driver) readLogs(info logger.Info, config logger.ReadConfig) (io.ReadCloser, error) {
	d.mu.Lock()
	l, ok := d.loggers[info.ContainerID]
	d.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("logger for %s not found", info.ContainerID)
	}

	lr, ok := l.(logger.LogReader)
	if !ok {
		return nil, fmt.Errorf("logger for %s not readable", info.ContainerID)
	}

	r, w := io.Pipe()
	go copyFromLogReader(lr, w)
	return r, nil
}

func (d *driver) capabilities() logger.Capability {
	return logger.Capability{ReadLogs: true}
}

func newLogger(info logger.Info) (logger.Logger, error) {
	return jsonfilelog.New(info)
}

func openFifo(file string) (io.ReadCloser, error) {
	return fifo.OpenFifo(context.Background(), file, syscall.O_RDONLY, 0)
}

func copyToLogger(r io.Reader, l logger.Logger) {
}

func copyFromLogReader(l logger.LogReader, w io.Writer) {
}
