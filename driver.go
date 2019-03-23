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
	"github.com/docker/docker/api/types/plugins/logdriver"
	"github.com/docker/docker/daemon/logger"
	"github.com/pkg/errors"
	"time"
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

	l, err := newTeeLogger(info)
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

	go doLog(r, l)
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
	go doReadLogs(lr, config, w)
	return r, nil
}

func (d *driver) capabilities() logger.Capability {
	return logger.Capability{ReadLogs: true}
}

func openFifo(file string) (io.ReadCloser, error) {
	return fifo.OpenFifo(context.Background(), file, syscall.O_RDONLY, 0)
}

func doLog(r io.ReadCloser, l logger.Logger) {
	defer r.Close()
	defer l.Close()

	dec := logdriver.NewLogEntryDecoder(r)
	var buf logdriver.LogEntry
	for {
		if err := dec.Decode(&buf); err != nil {
			if err == io.EOF {
				return
			}
			log.WithError(err).Error("failed to write log")
			return
		}
		msg := &logger.Message{
			Timestamp: time.Unix(0, buf.TimeNano),
			Line:      buf.Line,
			Source:    buf.Source,
		}
		l.Log(msg)
	}
}

func doReadLogs(lr logger.LogReader, config logger.ReadConfig, w io.WriteCloser) {
	defer w.Close()
	watcher := lr.ReadLogs(config)
	defer watcher.ConsumerGone()

	enc := logdriver.NewLogEntryEncoder(w)
	var buf logdriver.LogEntry
	for {
		select {
		case msg, ok := <-watcher.Msg:
			if !ok {
				log.Info("stop reading")
				return
			}
			buf.Line = msg.Line
			buf.TimeNano = msg.Timestamp.UnixNano()
			if msg.PLogMetaData != nil {
				buf.Partial = true
				buf.PartialLogMetadata = &logdriver.PartialLogEntryMetadata{
					Id:      msg.PLogMetaData.ID,
					Last:    msg.PLogMetaData.Last,
					Ordinal: int32(msg.PLogMetaData.Ordinal),
				}
			}
			buf.Source = msg.Source

			if err := enc.Encode(&buf); err != nil {
				log.WithError(err).Error("encode error")
				return
			}
			buf.Reset()
		case err := <-watcher.Err:
			log.WithError(err).Error("watcher error")
			return
		case <-watcher.WatchProducerGone():
			log.Info("producer gone")
			return
		}
	}
}
