package transport

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"nis-pipo/internal/meeting"
	"nis-pipo/internal/middleware"
)

type MeetingHandler struct {
	service *meeting.Service
}

func NewMeetingHandler(service *meeting.Service) *MeetingHandler {
	return &MeetingHandler{service: service}
}

type CreateMeetingRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DateStart   string `json:"date_start"`
	DateEnd     string `json:"date_end"`
	SlotMinutes int    `json:"slot_minutes"`
}

type UpdateMeetingRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type FinalizeMeetingRequest struct {
	FinalSlotIndex int `json:"final_slot_index"`
}

// ListMeetings godoc
//
//	@Summary	List meetings of current user (owner)
//	@Tags		meetings
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{array}		meeting.Meeting
//	@Failure	401	"unauthorized"
//	@Router		/api/meetings [get]
func (h *MeetingHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		ownerID, _ := r.Context().Value(middleware.UserIDKey).(string)
		if ownerID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		meetings, err := h.service.ListByOwner(r.Context(), ownerID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if meetings == nil {
			meetings = []meeting.Meeting{}
		}
		writeJSON(w, http.StatusOK, meetings)
	}
}

// CreateMeeting godoc
//
//	@Summary	Create meeting
//	@Tags		meetings
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		body	body		CreateMeetingRequest	true	"meeting data"
//	@Success	201		{object}	meeting.Meeting
//	@Failure	400		"bad request"
//	@Failure	401		"unauthorized"
//	@Router		/api/meetings [post]
func (h *MeetingHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		ownerID, _ := r.Context().Value(middleware.UserIDKey).(string)
		if ownerID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var req CreateMeetingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		dateStart, err := time.Parse("2006-01-02", req.DateStart)
		if err != nil {
			http.Error(w, "invalid date_start", http.StatusBadRequest)
			return
		}
		dateEnd, err := time.Parse("2006-01-02", req.DateEnd)
		if err != nil {
			http.Error(w, "invalid date_end", http.StatusBadRequest)
			return
		}
		m, err := h.service.Create(r.Context(), ownerID, req.Title, req.Description, dateStart, dateEnd, req.SlotMinutes)
		if err != nil {
			if err == meeting.ErrInvalidDates || err == meeting.ErrInvalidSlotMin {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, m)
	}
}

// GetMeeting godoc
//
//	@Summary	Get meeting by ID
//	@Tags		meetings
//	@Produce	json
//	@Param		id	path		string	true	"meeting ID"
//	@Success	200	{object}	meeting.Meeting
//	@Failure	404	"not found"
//	@Router		/api/meetings/{id} [get]
func (h *MeetingHandler) GetByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}
		m, err := h.service.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, m)
	}
}

// UpdateMeeting godoc
//
//	@Summary	Update meeting (title, description)
//	@Tags		meetings
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id		path		string				true	"meeting ID"
//	@Param		body	body		UpdateMeetingRequest	true	"title, description"
//	@Success	200		{object}	meeting.Meeting
//	@Failure	400		"bad request"
//	@Failure	401		"unauthorized"
//	@Failure	403		"forbidden"
//	@Failure	404		"not found"
//	@Router		/api/meetings/{id} [put]
func (h *MeetingHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		ownerID, _ := r.Context().Value(middleware.UserIDKey).(string)
		if ownerID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}
		var req UpdateMeetingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		m, err := h.service.Update(r.Context(), id, ownerID, req.Title, req.Description)
		if err != nil {
			if err == meeting.ErrNotFound {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			if err == meeting.ErrForbidden {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, m)
	}
}

// DeleteMeeting godoc
//
//	@Summary	Delete meeting
//	@Tags		meetings
//	@Security	BearerAuth
//	@Param		id	path		string	true	"meeting ID"
//	@Success	204	"no content"
//	@Failure	401	"unauthorized"
//	@Failure	403	"forbidden"
//	@Failure	404	"not found"
//	@Router		/api/meetings/{id} [delete]
func (h *MeetingHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		ownerID, _ := r.Context().Value(middleware.UserIDKey).(string)
		if ownerID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}
		err := h.service.Delete(r.Context(), id, ownerID)
		if err != nil {
			if err == meeting.ErrNotFound {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			if err == meeting.ErrForbidden {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// GetMeetingResults godoc
//
//	@Summary	Get meeting results (slot counts), owner only
//	@Tags		meetings
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		string	true	"meeting ID"
//	@Success	200	{array}		meeting.SlotResult
//	@Failure	401	"unauthorized"
//	@Failure	403	"forbidden"
//	@Failure	404	"not found"
//	@Router		/api/meetings/{id}/results [get]
func (h *MeetingHandler) GetResults() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		ownerID, _ := r.Context().Value(middleware.UserIDKey).(string)
		if ownerID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}
		results, err := h.service.GetResults(r.Context(), id, ownerID)
		if err != nil {
			if err == meeting.ErrNotFound {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			if err == meeting.ErrForbidden {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, results)
	}
}

// FinalizeMeeting godoc
//
//	@Summary	Finalize meeting (set final slot), owner only
//	@Tags		meetings
//	@Accept		json
//	@Security	BearerAuth
//	@Param		id		path		string					true	"meeting ID"
//	@Param		body	body		FinalizeMeetingRequest	true	"final_slot_index"
//	@Success	204		"no content"
//	@Failure	400		"bad request"
//	@Failure	401		"unauthorized"
//	@Failure	403		"forbidden"
//	@Failure	404		"not found"
//	@Router		/api/meetings/{id}/finalize [put]
func (h *MeetingHandler) Finalize() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		ownerID, _ := r.Context().Value(middleware.UserIDKey).(string)
		if ownerID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}
		var req FinalizeMeetingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		err := h.service.Finalize(r.Context(), id, ownerID, req.FinalSlotIndex)
		if err != nil {
			if err == meeting.ErrNotFound {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			if err == meeting.ErrForbidden {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			if err == meeting.ErrInvalidSlotIndex {
				http.Error(w, "final_slot_index out of range", http.StatusBadRequest)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

