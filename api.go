package main

import (
	"bytes"
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
	sdk.Handler
}

func newHandler() *handler {
	h := &handler{sdk.NewHandler(`{"Implements": ["LoggingDriver"]}`)}
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
	h.respond(nil, w)
}

func (h *handler) stopLogging(w http.ResponseWriter, r *http.Request) {
	var req stopLoggingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.WithField("file", req.File).Infof("stopLogging")
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

	stream := bytes.NewBuffer([]byte{})
	w.Header().Set("Content-Type", "application/x-json-stream")
	wf := ioutils.NewWriteFlusher(w)
	io.Copy(wf, stream)
}

func (h *handler) respond(err error, w http.ResponseWriter) {
	var res response
	if err != nil {
		res.Err = err.Error()
	}
	json.NewEncoder(w).Encode(&res)
}
