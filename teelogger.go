package main

import (
	"bytes"
	"errors"
	"github.com/docker/docker/daemon/logger"
	"strings"
)

var (
	errNoSuchDrivers = errors.New("no such drivers")
)

type teeLogger struct {
	loggers []logger.Logger
}

type multipleError struct {
	message string
	errs    []error
}

func newMultipleError(message string, errors []error) *multipleError {
	return &multipleError{message, errors}
}

func (e *multipleError) Error() string {
	buf := bytes.NewBufferString(e.message)
	for _, err := range e.errs {
		buf.WriteString("; " + err.Error())
	}
	return buf.String()
}

// --log-opt drivers=json-file,local --log-opt json-file:
func newTeeLogger(info logger.Info) (*teeLogger, error) {
	names, err := driverNames(info.Config)
	if err != nil {
		return nil, err
	}

	loggers := []logger.Logger{}
	closeLoggers := func() {
		for _, l := range loggers {
			l.Close()
		}
	}

	for _, name := range names {
		creator, err := logger.GetLogDriver(name)
		if err != nil {
			closeLoggers()
			return nil, err
		}

		newInfo := info
		newInfo.Config = driverConfig(name, info.Config)
		log.Infof("adding logger %s with config %v", name, newInfo.Config)

		l, err := creator(newInfo)
		if err != nil {
			log.WithError(err).Errorf("could not create logger %s", name)
			closeLoggers()
			return nil, err
		}
		loggers = append(loggers, l)
	}

	return &teeLogger{loggers}, nil
}

func driverNames(config map[string]string) ([]string, error) {
	if s, ok := config["drivers"]; ok {
		return strings.Split(s, ","), nil
	}
	return nil, errNoSuchDrivers
}

func driverConfig(driverName string, config map[string]string) map[string]string {
	newConfig := map[string]string{}
	for k, v := range config {
		ks := strings.SplitN(k, ":", 2)
		if len(ks) != 2 || ks[0] != driverName {
			continue
		}
		newConfig[ks[1]] = v
	}

	return newConfig
}

func (l *teeLogger) Log(msg *logger.Message) error {
	errs := []error{}
	for _, lg := range l.loggers {
		// copy message before logging to against resetting.
		// see https://github.com/moby/moby/blob/e4cc3adf81cc0810a416e2b8ce8eb4971e17a3a3/daemon/logger/logger.go#L40
		m := *msg
		if err := lg.Log(&m); err != nil {
			errs = append(errs, err)
		}
		// get message from pool to reduce message pool size
		logger.NewMessage()
	}
	if len(errs) != 0 {
		return newMultipleError("faild to log on some loggers", errs)
	}
	return nil
}

func (l *teeLogger) Name() string {
	return pluginName
}

func (l *teeLogger) Close() error {
	errs := []error{}
	for _, lg := range l.loggers {
		if err := lg.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return newMultipleError("faild to close on some loggers", errs)
	}
	return nil
}

func (l *teeLogger) ReadLogs(readConfig logger.ReadConfig) *logger.LogWatcher {
	for _, lg := range l.loggers {
		lr, ok := lg.(logger.LogReader)
		if ok {
			return lr.ReadLogs(readConfig)
		}
	}
	return logger.NewLogWatcher()
}
