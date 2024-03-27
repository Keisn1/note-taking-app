package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Keisn1/note-taking-app/domain"
	"github.com/go-chi/chi/v5"
)

type NotesCtrlr struct {
	NotesStore domain.NotesStore
	Logger     domain.Logger
}

func NewNotesCtrlr(store domain.NotesStore, logger domain.Logger) NotesCtrlr {
	return NotesCtrlr{NotesStore: store, Logger: logger}
}

var (
	ErrDB = errors.New("ErrDB")
)

func (nc *NotesCtrlr) Edit(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))

	if handleBadRequest(w, err, nc.Logger, "Edit", "userID") {
		return
	}

	noteID, err := strconv.Atoi(chi.URLParam(r, "noteID"))
	if handleBadRequest(w, err, nc.Logger, "Edit", "noteID") {
		return
	}

	var body map[string]string
	err = json.NewDecoder(r.Body).Decode(&body)
	if handleBadRequest(w, err, nc.Logger, "Edit", "json") {
		return
	}

	note, ok := body["note"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		nc.Logger.Errorf("ProcessAddNote invalid body: %v", err)
		return
	}

	err = nc.NotesStore.EditNote(userID, noteID, note)
	if handleError(w, err, http.StatusInternalServerError, nc.Logger, "Edit", "DBerror") {
		return
	}
	w.WriteHeader(http.StatusOK)
	nc.Logger.Infof("Success: Edit: userID %v noteID %v note %v", userID, noteID, note)
}

func (nc *NotesCtrlr) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if handleBadRequest(w, err, nc.Logger, "Delete", "userID") {
		return
	}

	noteID, err := strconv.Atoi(chi.URLParam(r, "noteID"))
	if handleBadRequest(w, err, nc.Logger, "Delete", "noteID") {
		return
	}

	err = nc.NotesStore.Delete(userID, noteID)
	if handleError(w, err, http.StatusInternalServerError, nc.Logger, "Delete", "DBerror") {
		return
	}

	w.WriteHeader(http.StatusNoContent)
	nc.Logger.Infof("Success: Delete noteID %v userID %v", noteID, userID)
}

func (nc *NotesCtrlr) Add(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if handleBadRequest(w, err, nc.Logger, "ProcessAddNote", "userID") {
		return
	}

	var body map[string]string
	err = json.NewDecoder(r.Body).Decode(&body)
	if handleBadRequest(w, err, nc.Logger, "ProcessAddNote", "json") {
		return
	}

	note, ok := body["note"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		nc.Logger.Errorf("ProcessAddNote invalid body: %w", err)
		return
	}

	err = nc.NotesStore.AddNote(userID, note)
	if handleError(w, err, http.StatusConflict, nc.Logger, "ProcessAddNote", "DBerror") {
		return
	}

	w.WriteHeader(http.StatusAccepted)
	nc.Logger.Infof("Success: ProcessAddNote with userID %d and note %v", userID, note)
}

func (nc *NotesCtrlr) GetNotesByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if handleBadRequest(w, err, nc.Logger, "GetNotesByUserID", "userID") {
		return
	}

	notes, err := nc.NotesStore.GetNotesByUserID(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		nc.Logger.Errorf("GetNotesByUserID userID %v %v: %w", userID, ErrDB.Error(), err)
		return
	}

	err = json.NewEncoder(w).Encode(notes)
	if handleError(w, err, http.StatusInternalServerError, nc.Logger, "GetNotesByUserID", "invalid json") {
		return
	}

	nc.Logger.Infof("Success: GetNotesByUserID with userID %d", userID)
}

func (nc *NotesCtrlr) GetNoteByUserIDAndNoteID(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil || userID < 0 {
		http.Error(w, "", http.StatusBadRequest)
		nc.Logger.Errorf("GetNoteByUserIDandNoteID: invalid userID %v", userID)
		return
	}

	noteID, err := strconv.Atoi(chi.URLParam(r, "noteID"))
	if err != nil || noteID < 0 {
		http.Error(w, "", http.StatusBadRequest)
		nc.Logger.Errorf("GetNoteByUserIDandNoteID: invalid noteID %v", noteID)
		return
	}

	notes, err := nc.NotesStore.GetNoteByUserIDAndNoteID(userID, noteID)
	if err != nil {
		http.Error(w, "", http.StatusNotFound)
		msg := fmt.Sprintf("GetNoteByUserIDAndNoteID userID %v and noteID %v", userID, noteID)
		nc.Logger.Errorf("%s: %w", msg, err)
		return
	}

	err = json.NewEncoder(w).Encode(notes)
	if handleError(w, err, http.StatusInternalServerError, nc.Logger, "GetNotesByUserID", "invalid json") {
		return
	}

	nc.Logger.Infof("Success: GetNoteByUserIDAndNoteID with userID %v and noteID %v", userID, noteID)
}

func (nc *NotesCtrlr) GetAllNotes(w http.ResponseWriter, r *http.Request) {
	notes, err := nc.NotesStore.GetAllNotes()
	if handleError2(w, err, http.StatusInternalServerError, nc.Logger, "GetAllNotes") {
		return
	}

	err = json.NewEncoder(w).Encode(notes)
	if handleError(w, err, http.StatusInternalServerError, nc.Logger, "GetAllNotes", "invalid json") {
		return
	}

	nc.Logger.Infof("Success: GetAllNotes")
}

func handleBadRequest(w http.ResponseWriter, err error, logger domain.Logger, action, param string) bool {
	if err != nil {
		logger.Errorf("%s invalid %s: %v", action, param, err)
		http.Error(w, "", http.StatusBadRequest)
		return true
	}
	return false
}

func handleError(w http.ResponseWriter, err error, httpErr int, logger domain.Logger, action, msg string) bool {
	if err != nil {
		logger.Errorf("%s: %s: %w", action, msg, err)
		http.Error(w, "", httpErr)
		return true
	}
	return false
}

func handleError2(w http.ResponseWriter, err error, httpErr int, logger domain.Logger, handler string) bool {
	if err != nil {
		logger.Errorf("%s: %w", handler, err)
		http.Error(w, "", httpErr)
		return true
	}
	return false
}
