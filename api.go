package main

import (
	"encoding/json"
	"io"
	"net/http"

	"context"
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
	if err := sdk.DecodeRequest(w, r, &req); err != nil {
		return
	}

	log.WithField("file", req.File).WithField("containerID", req.Info.ContainerID).Infof("startLogging")
	err := h.d.startLogging(req.File, req.Info)
	h.encodeResponse(w, err)
}

func (h *handler) stopLogging(w http.ResponseWriter, r *http.Request) {
	var req stopLoggingRequest
	if err := sdk.DecodeRequest(w, r, &req); err != nil {
		return
	}

	log.WithField("file", req.File).Infof("stopLogging")
	err := h.d.stopLogging(req.File)
	h.encodeResponse(w, err)
}

func (h *handler) capabilities(w http.ResponseWriter, r *http.Request) {
	log.Infof("capabilities")
	json.NewEncoder(w).Encode(&capabilitiesResponse{
		Cap: h.d.capabilities(),
	})
}

func (h *handler) readLogs(w http.ResponseWriter, r *http.Request) {
	var req readLogsRequest
	if err := sdk.DecodeRequest(w, r, &req); err != nil {
		return
	}

	log.WithField("ContainerID", req.Info.ContainerID).Infof("readLogs")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := h.d.readLogs(ctx, req.Info, req.Config)

	if err != nil {
		h.encodeResponse(w, err)
	} else {
		if req.Config.Follow {
			defer stream.Close()
			// Don't use sdk.StreamResponse for flushing.
			w.Header().Set("Content-Type", sdk.DefaultContentTypeV1_1)
			io.Copy(ioutils.NewWriteFlusher(w), stream)
		} else {
			sdk.StreamResponse(w, stream)
		}
	}
}

func (h *handler) encodeResponse(w http.ResponseWriter, err error) {
	if err != nil {
		sdk.EncodeResponse(w, &response{Err: err.Error()}, true)
	} else {
		sdk.EncodeResponse(w, &response{Err: ""}, false)
	}
}
