package main

import (
	"encoding/json"
	"net/http"

	"github.com/docker/docker/daemon/logger"
	"github.com/docker/go-plugins-helpers/sdk"
)

type startLoggingRequest struct {
	File string
	Info logger.Info
}

type stopLoggingRequest struct {
	File string
}

type capabilitiesResponse struct {
	Err string
	Cap logger.Capability
}

type readLogsRequest struct {
	Info   logger.Info
	Config logger.ReadConfig
}

type response struct {
	Err string
}

type handler struct {
	d *driver
	sdk.Handler
}

func newHandler(d *driver) *handler {
	h := &handler{d, sdk.NewHandler(`{"Implements": ["LoggingDriver"]}`)}
	h.HandleFunc("/LogDriver.StartLogging", h.startLogging)
	h.HandleFunc("/LogDriver.StopLogging", h.stopLogging)
	h.HandleFunc("/LogDriver.Capabilities", h.capabilities)
	h.HandleFunc("/LogDriver.ReadLogs", h.readLogs)
	return h
}

func (h *handler) startLogging(w http.ResponseWriter, r *http.Request) {
	var req startLoggingRequest
	if err := sdk.DecodeRequest(w, r, &req); err != nil {
		return
	}

	log.WithField("file", req.File).WithField("containerID", req.Info.ContainerID).Infof("startLogging")
	h.d.startLogging(req.File, req.Info)
	sdk.EncodeResponse(w, nil, false)
}

func (h *handler) stopLogging(w http.ResponseWriter, r *http.Request) {
	var req stopLoggingRequest
	if err := sdk.DecodeRequest(w, r, &req); err != nil {
		return
	}

	log.WithField("file", req.File).Infof("stopLogging")
	h.d.stopLogging(req.File)
	sdk.EncodeResponse(w, nil, false)
}

func (h *handler) capabilities(w http.ResponseWriter, r *http.Request) {
	log.Infof("capabilities")
	json.NewEncoder(w).Encode(&capabilitiesResponse{
		Cap: logger.Capability{ReadLogs: true},
	})
}

func (h *handler) readLogs(w http.ResponseWriter, r *http.Request) {
	var req readLogsRequest
	if err := sdk.DecodeRequest(w, r, &req); err != nil {
		return
	}

	log.WithField("ContainerID", req.Info.ContainerID).Infof("readLogs")
	stream := h.d.readLogs(req.Info, req.Config)
	sdk.StreamResponse(w, stream)
}
