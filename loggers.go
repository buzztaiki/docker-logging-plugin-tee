package main

import (
	"github.com/docker/docker/daemon/logger"
	"github.com/docker/docker/pkg/plugingetter"
	"github.com/docker/docker/pkg/plugins"
	"github.com/pkg/errors"

	_ "github.com/docker/docker/daemon/logger/awslogs"
	_ "github.com/docker/docker/daemon/logger/fluentd"
	_ "github.com/docker/docker/daemon/logger/gcplogs"
	_ "github.com/docker/docker/daemon/logger/gelf"
	_ "github.com/docker/docker/daemon/logger/journald"
	_ "github.com/docker/docker/daemon/logger/jsonfilelog"
	_ "github.com/docker/docker/daemon/logger/local"
	_ "github.com/docker/docker/daemon/logger/logentries"
	_ "github.com/docker/docker/daemon/logger/splunk"
	_ "github.com/docker/docker/daemon/logger/syslog"
)

func init() {
	logger.RegisterPluginGetter(&nullPluginGetter{})
}

type nullPluginGetter struct{}

var errNullPluginGetterGet error = errors.New("failed to get")

func (pg *nullPluginGetter) Get(name, capability string, mode int) (plugingetter.CompatPlugin, error) {
	return nil, errNullPluginGetterGet
}

func (pg *nullPluginGetter) GetAllByCap(capability string) ([]plugingetter.CompatPlugin, error) {
	return nil, errNullPluginGetterGet
}

func (pg *nullPluginGetter) GetAllManagedPluginsByCap(capability string) []plugingetter.CompatPlugin {
	return nil
}

func (pg *nullPluginGetter) Handle(capability string, callback func(string, *plugins.Client)) {
}
