package http

import (
	"encoding/json"
	"net/http"

	"messaging/server/internal/domain"
	"messaging/server/internal/usecase"
	"messaging/server/internal/usecase/port"
)

type Handler struct {
	RegisterDevice *usecase.RegisterDevice
	DeviceRepo     port.DeviceRepo
}

func NewHandler(rd *usecase.RegisterDevice, repo port.DeviceRepo) *Handler {
	return &Handler{
		RegisterDevice: rd,
		DeviceRepo:     repo,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /devices", h.handleCreateDevice)
	mux.HandleFunc("GET /devices", h.handleListDevices)
	mux.HandleFunc("DELETE /devices/{id}", h.handleDeleteDevice)
}

func (h *Handler) handleCreateDevice(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name   string `json:"name"`
		Avatar string `json:"avatar"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	out, err := h.RegisterDevice.Execute(r.Context(), usecase.RegisterDeviceInput{
		Name:   body.Name,
		Avatar: body.Avatar,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(out.Device)
}

func (h *Handler) handleListDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := h.DeviceRepo.ListAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// If devices is nil, we want to return [] rather than null in JSON
	if devices == nil {
		devices = []domain.Device{}
	}
	json.NewEncoder(w).Encode(devices)
}

func (h *Handler) handleDeleteDevice(w http.ResponseWriter, r *http.Request) {
	id := domain.DeviceID(r.PathValue("id"))
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	if err := h.DeviceRepo.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
