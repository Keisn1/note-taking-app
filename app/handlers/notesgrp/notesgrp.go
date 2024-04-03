package notesgrp

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Keisn1/note-taking-app/domain"
	"github.com/Keisn1/note-taking-app/foundation"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handlers struct {
	NotesStore domain.NotesStore
}

func NewHandlers(store domain.NotesStore) Handlers {
	return Handlers{NotesStore: store}
}

func (nc *Handlers) Edit(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(foundation.UserIDKey).(uuid.UUID)
	if !ok {
		logMsg := "Edit: invalid userID"
		handleError(w, "", http.StatusBadRequest, logMsg)
		return
	}

	noteID, err := strconv.Atoi(chi.URLParam(r, "noteID"))
	if err != nil || noteID < 0 {
		logMsg := fmt.Sprintf("Edit: invalid noteID %v", chi.URLParam(r, "noteID"))
		handleError(w, "", http.StatusBadRequest, logMsg, "error", err)
		return
	}

	var np domain.NotePost
	err = json.NewDecoder(r.Body).Decode(&np)
	if err != nil {
		logMsg := "Add: invalid body:"
		handleError(w, "", http.StatusBadRequest, logMsg, "error", err)
		return
	}

	err = nc.NotesStore.EditNote(userID, noteID, np.Note)
	if err != nil {
		logMsg := fmt.Sprintf("Edit: userID %v noteID %v body %v", userID, noteID, np)
		handleError(w, "", http.StatusConflict, logMsg, "error", err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	slog.Info(
		fmt.Sprintf("Success: Edit: userID %v noteID %v body %v", userID, noteID, np),
	)
}

func (nc *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(foundation.UserIDKey).(uuid.UUID)
	if !ok {
		logMsg := "Delete: invalid userID"
		handleError(w, "", http.StatusBadRequest, logMsg)
		return
	}

	noteID, err := strconv.Atoi(chi.URLParam(r, "noteID"))
	if err != nil || noteID < 0 {
		logMsg := fmt.Sprintf("Delete: invalid noteID %v", chi.URLParam(r, "noteID"))
		handleError(w, "", http.StatusBadRequest, logMsg, "error", err)
		return
	}

	err = nc.NotesStore.Delete(userID, noteID)
	if err != nil {
		logMsg := fmt.Sprintf("Delete: userID %v and noteID %v", userID, noteID)
		handleError(w, "", http.StatusNotFound, logMsg, "error", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	slog.Info(fmt.Sprintf("Success: Delete: userID %v noteID %v", userID, noteID))
}

func (nc *Handlers) Add(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(foundation.UserIDKey).(uuid.UUID)
	if !ok {
		logMsg := "Add: invalid userID"
		handleError(w, "", http.StatusBadRequest, logMsg)
		return
	}

	var np domain.NotePost
	err := json.NewDecoder(r.Body).Decode(&np)
	if err != nil {
		logMsg := "Add: invalid body"
		handleError(w, "", http.StatusBadRequest, logMsg, "error", err)
		return
	}

	err = nc.NotesStore.AddNote(userID, np.Note)
	if err != nil {
		logMsg := fmt.Sprintf("Add: userID %v body %v", userID, np)
		handleError(w, "", http.StatusConflict, logMsg, "error", err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	slog.Info(
		fmt.Sprintf("Success: Add: userID %v body %v", userID, np),
	)
}

func (nc *Handlers) GetNotesByUserID(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(foundation.UserIDKey).(uuid.UUID)
	if !ok {
		logMsg := "GetNotesByUserID: invalid userID"
		handleError(w, "", http.StatusBadRequest, logMsg)
		return
	}

	notes, err := nc.NotesStore.GetNotesByUserID(userID)
	if err != nil {
		logMsg := fmt.Sprintf("GetNotesByUserID: userID %v", userID)
		handleError(w, "", http.StatusInternalServerError, logMsg, "error", err)
		return
	}

	err = json.NewEncoder(w).Encode(notes)
	if err != nil {
		logMsg := fmt.Sprintf("GetNotesByUserID: userID %v: json encoding error", userID)
		handleError(w, "", http.StatusInternalServerError, logMsg, "error", err)
		return
	}

	slog.Info(fmt.Sprintf("Success: GetNotesByUserID: userID %v", userID))
}

func (nc *Handlers) GetNoteByUserIDAndNoteID(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(foundation.UserIDKey).(uuid.UUID)
	if !ok {
		logMsg := "GetNoteByUserIDandNoteID: invalid userID"
		handleError(w, "", http.StatusBadRequest, logMsg)
		return
	}

	noteID, err := strconv.Atoi(chi.URLParam(r, "noteID"))
	if err != nil || noteID < 0 {
		logMsg := fmt.Sprintf("GetNoteByUserIDandNoteID: invalid noteID %v", chi.URLParam(r, "noteID"))
		handleError(w, "", http.StatusBadRequest, logMsg, "error", err)
		return
	}

	notes, err := nc.NotesStore.GetNoteByUserIDAndNoteID(userID, noteID)
	if err != nil {
		logMsg := fmt.Sprintf("GetNoteByUserIDAndNoteID: userID %v noteID %v", userID, noteID)
		handleError(w, "", http.StatusInternalServerError, logMsg, "error", err)
		return
	}

	if len(notes) == 0 {
		logMsg := fmt.Sprintf("GetNoteByUserIDAndNoteID: userID %v noteID %v", userID, noteID)
		handleError(w, "", http.StatusNotFound, logMsg, "error", "Not Found")
		return
	}

	err = json.NewEncoder(w).Encode(notes)
	if err != nil {
		logMsg := fmt.Sprintf("GetNoteByUserIDAndNoteID: userID %v noteID %v: json encoding error", userID, noteID)
		handleError(w, "", http.StatusInternalServerError, logMsg, "error", err)
		return
	}

	slog.Info(fmt.Sprintf(
		"Success: GetNoteByUserIDAndNoteID: userID %v noteID %v", userID, noteID,
	))
}

func (nc *Handlers) GetAllNotes(w http.ResponseWriter, r *http.Request) {
	notes, err := nc.NotesStore.GetAllNotes()
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		slog.Error("GetAllNotes: DBError", "error", err)
		return
	}

	err = json.NewEncoder(w).Encode(notes)
	if err != nil {
		logMsg := "GetAllNotes: json encoding error"
		handleError(w, "", http.StatusInternalServerError, logMsg, "error", err)
		return
	}

	slog.Info("Success: GetAllNotes")
}

func handleError(w http.ResponseWriter, errMsg string, status int, logMsg string, args ...any) {
	http.Error(w, "", status)
	slog.Error(logMsg, args...)
}