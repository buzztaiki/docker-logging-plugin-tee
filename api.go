package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/docker/docker/daemon/logger"
	"github.com/docker/docker/pkg/ioutils"
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.WithField("file", req.File).WithField("containerID", req.Info.ContainerID).Infof("startLogging")
	h.d.startLogging(req.File, req.Info)
	h.respond(nil, w)
}

func (h *handler) stopLogging(w http.ResponseWriter, r *http.Request) {
	var req stopLoggingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.WithField("file", req.File).Infof("stopLogging")
	h.d.stopLogging(req.File)
	h.respond(nil, w)
}

func (h *handler) capabilities(w http.ResponseWriter, r *http.Request) {
	log.Infof("capabilities")
	json.NewEncoder(w).Encode(&capabilitiesResponse{
		Cap: logger.Capability{ReadLogs: true},
	})
}

func (h *handler) readLogs(w http.ResponseWriter, r *http.Request) {
	var req readLogsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.WithField("ContainerID", req.Info.ContainerID).Infof("readLogs")

	w.Header().Set("Content-Type", "application/x-json-stream")
	stream := h.d.readLogs(req.Info, req.Config)
	io.Copy(ioutils.NewWriteFlusher(w), stream)
}

func (h *handler) respond(err error, w http.ResponseWriter) {
	var res response
	if err != nil {
		res.Err = err.Error()
	}
	json.NewEncoder(w).Encode(&res)
}
