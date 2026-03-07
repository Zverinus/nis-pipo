package transport

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nis-pipo/internal/participant"
)

type ParticipantHandler struct {
	service *participant.Service
}

func NewParticipantHandler(service *participant.Service) *ParticipantHandler {
	return &ParticipantHandler{service: service}
}

type CreateParticipantRequest struct {
	DisplayName string `json:"display_name"`
}

type SetSlotsRequest struct {
	SlotIndexes []int `json:"slot_indexes"`
}

// CreateParticipant godoc
//
//	@Summary	Create participant for meeting
//	@Tags		participants
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string						true	"meeting ID"
//	@Param		body	body		CreateParticipantRequest		true	"display_name"
//	@Success	201		{object}	map[string]string
//	@Failure	400		"bad request"
//	@Failure	404		"meeting not found"
//	@Failure	409		"meeting finalized"
//	@Router		/api/meetings/{id}/participants [post]
func (h *ParticipantHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		meetingID := chi.URLParam(r, "id")
		if meetingID == "" {
			http.Error(w, "meeting id required", http.StatusBadRequest)
			return
		}
		var req CreateParticipantRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		p, err := h.service.Create(r.Context(), meetingID, req.DisplayName)
		if err != nil {
			if err == participant.ErrMeetingNotFound {
				http.Error(w, "meeting not found", http.StatusNotFound)
				return
			}
			if err == participant.ErrMeetingFinalized {
				http.Error(w, "meeting is finalized", http.StatusConflict)
				return
			}
			logError(r, "create participant", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"id": p.ID})
	}
}

// SetParticipantSlots godoc
//
//	@Summary	Set participant slot choices
//	@Tags		participants
//	@Accept		json
//	@Param		id				path		string				true	"meeting ID"
//	@Param		participant_id	path		string				true	"participant ID"
//	@Param		body			body		SetSlotsRequest		true	"slot_indexes"
//	@Success	204		"no content"
//	@Failure	400		"bad request"
//	@Failure	404		"participant or meeting not found"
//	@Failure	409		"meeting finalized"
//	@Router		/api/meetings/{id}/participants/{participant_id}/slots [put]
func (h *ParticipantHandler) SetSlots() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		meetingID := chi.URLParam(r, "id")
		participantID := chi.URLParam(r, "participant_id")
		if meetingID == "" || participantID == "" {
			http.Error(w, "meeting id and participant id required", http.StatusBadRequest)
			return
		}
		var req SetSlotsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		err := h.service.SetSlots(r.Context(), meetingID, participantID, req.SlotIndexes)
		if err != nil {
			if err == participant.ErrMeetingNotFound || err == participant.ErrParticipantNotFound {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			if err == participant.ErrMeetingFinalized {
				http.Error(w, "meeting is finalized", http.StatusConflict)
				return
			}
			if err == participant.ErrSlotOutOfRange {
				http.Error(w, "slot_index out of range", http.StatusBadRequest)
				return
			}
			logError(r, "set slots", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
