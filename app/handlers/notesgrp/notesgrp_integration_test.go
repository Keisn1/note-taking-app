package notesgrp_test

// import (
// 	"bytes"
// 	"encoding/json"
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"strconv"
// 	"testing"

// 	"fmt"

// 	"github.com/Keisn1/note-taking-app/app/api"
// 	"github.com/Keisn1/note-taking-app/app/handlers/notesgrp"
// 	"github.com/Keisn1/note-taking-app/domain"
// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// )

// func allNotes() domain.Notes {
// 	uid1 := uuid.UUID([16]byte{1})
// 	uid2 := uuid.UUID([16]byte{2})
// 	return domain.Notes{
// 		{UserID: uid1, Note: "Test note 1"},
// 		{UserID: uid1, Note: "Test note 2"},
// 		{UserID: uid1, Note: "Test note 3"},
// 		{UserID: uid2, Note: "Test note 4"},
// 		{UserID: uid2, Note: "Test note 5"},
// 		{UserID: uid2, Note: "Test note 6"},
// 	}
// }

// func TestIntegration(t *testing.T) {
// 	store := notesgrp.NewInMemoryNotesStore()
// 	notesC := notesgrp.NewHandlers(store)

// 	// Add notes
// 	addNotes(t, notesC)

// 	// Testing all notes
// 	canRetrieveAllNotes(t, notesC)

// 	// Testing notes by userID
// 	canRetrieveNotesByUserID(t, notesC)

// 	// Testing notes by userID and noteID
// 	canRetrieveNotesByUserIDAndNoteID(t, notesC)

// 	// Edit a note
// 	canEditNote(t, notesC)

// 	// Delete a note
// 	canDeleteNotes(t, notesC)
// }

// func retrieveAllNotes(t *testing.T, hdl notesgrp.Handlers) domain.Notes {
// 	rr := httptest.NewRecorder()
// 	req := setupRequest(t, "GET", "/notes", nil, "", &bytes.Buffer{})
// 	hdl.GetAllNotes(rr, req)
// 	return decodeBodyNotes(t, rr.Body)
// }

// func canRetrieveAllNotes(t *testing.T, hdl notesgrp.Handlers) {
// 	t.Helper()
// 	gotNotes := retrieveAllNotes(t, hdl)
// 	gotNotes = setNoteIdZero(gotNotes)
// 	assert.Equal(t, allNotes(), gotNotes)
// }

// func canRetrieveNotesByUserID(t *testing.T, hdl notesgrp.Handlers) {
// 	t.Helper()
// 	uid1 := uuid.UUID([16]byte{1})
// 	uid2 := uuid.UUID([16]byte{2})
// 	for _, uid := range []uuid.UUID{uid1, uid2} {
// 		var wantNotes domain.Notes
// 		for _, n := range allNotes() {
// 			if n.UserID == uid {
// 				wantNotes = append(wantNotes, n)
// 			}
// 		}

// 		rr := httptest.NewRecorder()
// 		req := setupRequest(t, "GET", "/users/notes", uid, "", &bytes.Buffer{})
// 		hdl.GetNotesByUserID(rr, req)
// 		gotNotes := decodeBodyNotes(t, rr.Body)
// 		gotNotes = setNoteIdZero(gotNotes)

// 		assert.Equal(t, http.StatusOK, rr.Code)
// 		assert.Equal(t, wantNotes, gotNotes)
// 	}
// }

// func canRetrieveNotesByUserIDAndNoteID(t *testing.T, hdl notesgrp.Handlers) {
// 	t.Helper()
// 	allNotes := retrieveAllNotes(t, hdl)

// 	for _, n := range allNotes {
// 		rr := httptest.NewRecorder()
// 		req := setupRequest(t, "POST", "/users/notes/{noteID}", n.UserID, strconv.Itoa(n.NoteID), &bytes.Buffer{})
// 		hdl.GetNoteByUserIDAndNoteID(rr, req)
// 		gotNotes := decodeBodyNotes(t, rr.Body)

// 		assert.Equal(t, http.StatusOK, rr.Code)
// 		assert.Equal(t, domain.Notes{n}, gotNotes)
// 	}
// }

// func canDeleteNotes(t *testing.T, hdl notesgrp.Handlers) {
// 	t.Helper()

// 	allNotes := retrieveAllNotes(t, hdl)
// 	for _, n := range allNotes {
// 		rr := httptest.NewRecorder()
// 		req := setupRequest(t, "DELETE", "/users/notes/{noteID}", n.UserID, strconv.Itoa(n.NoteID), &bytes.Buffer{})
// 		hdl.Delete(rr, req)
// 		assert.Equal(t, http.StatusNoContent, rr.Code)

// 		rr = httptest.NewRecorder()
// 		req = setupRequest(t, "GET", "/users/notes/{noteID}", n.UserID, strconv.Itoa(n.NoteID), &bytes.Buffer{})
// 		hdl.GetNoteByUserIDAndNoteID(rr, req)

// 		// gotNotes := decodeBodyNotes(t, rr.Body)
// 		assert.Equal(t, http.StatusNotFound, rr.Code)
// 	}

// 	allNotes = retrieveAllNotes(t, hdl)
// 	assert.Equal(t, 0, len(allNotes))
// }

// func canEditNote(t *testing.T, hdl notesgrp.Handlers) {
// 	allNotes := retrieveAllNotes(t, hdl)

// 	for _, n := range allNotes {
// 		n.Note = fmt.Sprintf("Edited note userID %v noteID %v", n.UserID, n.NoteID)

// 		// edit note
// 		body := api.NotePost{Note: n.Note}
// 		rr := httptest.NewRecorder()
// 		req := setupRequest(t, "POST", "/users/notes/{noteID}", n.UserID, strconv.Itoa(n.NoteID), mustEncode(t, body))
// 		hdl.Edit(rr, req)
// 		assert.Equal(t, http.StatusAccepted, rr.Code)

// 		// test
// 		rr = httptest.NewRecorder()
// 		req = setupRequest(t, "GET", "/users/notes/{noteID}", n.UserID, strconv.Itoa(n.NoteID), &bytes.Buffer{})
// 		hdl.GetNoteByUserIDAndNoteID(rr, req)
// 		gotNotes := decodeBodyNotes(t, rr.Body)
// 		assert.Equal(t, domain.Notes{n}, gotNotes)
// 	}
// }
// func setNoteIdZero(notes domain.Notes) domain.Notes {
// 	var newNotes domain.Notes
// 	for _, n := range notes {
// 		n.NoteID = 0
// 		newNotes = append(newNotes, n)
// 	}
// 	return newNotes
// }

// func decodeBodyNotes(t testing.TB, body io.Reader) (notes domain.Notes) {
// 	t.Helper()
// 	err := json.NewDecoder(body).Decode(&notes)
// 	if err != nil {
// 		t.Fatalf("Unable to parse body into Notes: %v", err)
// 	}
// 	return
// }

// func addNotes(t *testing.T, hdl notesgrp.Handlers) {
// 	t.Helper()
// 	for _, n := range allNotes() {
// 		body := api.NotePost{Note: n.Note}
// 		req := setupRequest(t, "POST", "/users/notes", n.UserID, strconv.Itoa(n.NoteID), mustEncode(t, body))
// 		hdl.Add(
// 			httptest.NewRecorder(),
// 			req,
// 		)
// 	}
// }
