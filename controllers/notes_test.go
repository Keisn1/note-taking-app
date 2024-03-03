package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-chi/chi"
)

func TestNotes(t *testing.T) {
	notesStore := NewStubNotesStore()
	logger := NewStubLogger()
	notesC := NewNotesCtrlr(notesStore, logger)

	t.Run("Server returns all Notes", func(t *testing.T) {
		logger.Reset()

		request := newGetAllNotesRequest(t)
		response := httptest.NewRecorder()
		notesC.GetAllNotes(response, request)

		assertStatusCode(t, response.Result().StatusCode, http.StatusOK)
		assertAllNotesGotCalled(t, notesStore.allNotesGotCalled)
		assertLoggingCalls(t, logger.infofCalls, []fmtCallf{
			{format: "Success: GetAllNotes"},
		})
	})

	t.Run("Return notes for user with userID", func(t *testing.T) {
		logger.Reset()
		testCases := []struct {
			userID     int
			statusCode int
		}{
			{1, http.StatusOK},
			{2, http.StatusOK},
			{100, http.StatusNotFound},
		}

		for _, tc := range testCases {
			response := httptest.NewRecorder()
			request := newGetNotesByUserIdRequest(t, tc.userID)
			notesC.GetNotesByUserID(response, request)
			assertStatusCode(t, response.Result().StatusCode, tc.statusCode)
		}
		assertEqualIntSlice(t, notesStore.getNotesByUserIDCalls, []int{1, 2, 100})
		assertLoggingCalls(t, logger.infofCalls, []fmtCallf{
			{format: "Success: GetNotesByUserID with userID %d", a: []any{1}},
			{format: "Success: GetNotesByUserID with userID %d", a: []any{2}},
		})
		assertLoggingCalls(t, logger.errorfCall, []fmtCallf{
			{format: "Failure: GetNotesByUserID with userID %d", a: []any{100}},
		})

	})

	t.Run("test false url parameters throw error", func(t *testing.T) {
		logger.Reset()

		badID := "notAnInt"
		badRequest := newRequestWithBadIdParam(t, badID)
		response := httptest.NewRecorder()
		notesC.GetNotesByUserID(response, badRequest)

		assertStatusCode(t, response.Result().StatusCode, http.StatusBadRequest)
		assertLoggingCalls(t, logger.errorfCall, []fmtCallf{
			{format: "Failure: GetNotesByUserID: %w: %v", a: []any{ErrInvalidUserID, badID}},
		})
	})

	t.Run("POST a Note", func(t *testing.T) {
		logger.Reset()
		userID, note := 1, "Test note"

		request := newPostRequestWithNote(t, note)
		request = WithUrlParam(request, "userID", fmt.Sprintf("%d", userID))
		response := httptest.NewRecorder()
		notesC.ProcessAddNote(response, request)

		wantAddNoteCalls := []AddNoteCall{{userID: userID, note: note}}
		assertStatusCode(t, response.Result().StatusCode, http.StatusAccepted)
		assertAddNoteCallsEqual(t, notesStore.addNoteCalls, wantAddNoteCalls)
		assertLoggingCalls(t, logger.infofCalls, []fmtCallf{
			{format: "Success: ProcessAddNote with userID %d and note %v", a: []any{userID, note}},
		})
	})

	t.Run("test invalid json body", func(t *testing.T) {
		logger.Reset()
		badRequest := newPostRequestFromBody(t, "{}}", "")
		badRequest = WithUrlParam(badRequest, "userID", fmt.Sprintf("%d", 1))
		response := httptest.NewRecorder()
		notesC.ProcessAddNote(response, badRequest)

		assertStatusCode(t, response.Result().StatusCode, http.StatusBadRequest)
		assertLoggingCalls(t, logger.errorfCall, []fmtCallf{
			{format: "Failure: ProcessAddNote: %w: %v", a: []any{ErrUnmarshalRequestBody, badRequest.Body}},
		})
	})

	t.Run("test invalid request body", func(t *testing.T) {
		logger.Reset()

		badRequest := newInvalidBodyPostRequest(t)
		badRequest = WithUrlParam(badRequest, "userID", fmt.Sprintf("%d", 1))
		response := httptest.NewRecorder()
		notesC.ProcessAddNote(response, badRequest)

		assertStatusCode(t, response.Result().StatusCode, http.StatusBadRequest)
		assertLoggingCalls(t, logger.errorfCall, []fmtCallf{
			{format: "Failure: ProcessAddNote: %w: %v", a: []any{ErrInvalidRequestBody, badRequest.Body}},
		})
	})

	// t.Run("test AddNote and Note already present", func(t *testing.T) {
	// 	logger.Reset()

	// 	request := newPostRequestWithNote(t, NewNote(1, "Note already present"), "/notes/1")
	// 	response := httptest.NewRecorder()

	// 	notesC.ProcessAddNote(response, request)
	// 	assertStatusCode(t, response.Result().StatusCode, http.StatusInternalServerError)
	// 	assertLoggingCalls(t, logger.errorfCall, []fmtCallf{
	// 		{format: "Failure: ProcessAddNote: %w", a: []any{ErrDBResourceCreation}},
	// 	})
	// })

	// t.Run("Delete a Note", func(t *testing.T) {
	// 	logger.Reset()

	// 	deleteRequest := newDeleteRequest(t, 1)
	// 	response := httptest.NewRecorder()
	// 	notesC.Delete(response, deleteRequest)

	// 	wantDeleteNoteCalls := []int{1}
	// 	assertStatusCode(t, response.Result().StatusCode, http.StatusNoContent)
	// 	assertEqualIntSlice(t, notesStore.deleteNoteCalls, wantDeleteNoteCalls)
	// })

	// t.Run("Deletion fail", func(t *testing.T) {
	// 	logger.Reset()

	// 	deleteRequest := newDeleteRequest(t, 50) // id not present
	// 	response := httptest.NewRecorder()
	// 	notesC.Delete(response, deleteRequest)

	// 	assertStatusCode(t, response.Result().StatusCode, http.StatusNotFound)
	// 	assertRightErrorCall(t, logger.errorfCall[0], "%w: %w", ErrDBResourceDeletion)
	// })

	// t.Run("Edit a Note", func(t *testing.T) {
	// 	logger.Reset()

	// 	note := NewNote(1, "Edited note")
	// 	putRequest := newPutRequestWithNote(t, note, "/notes/1")
	// 	response := httptest.NewRecorder()
	// 	notesC.Edit(response, putRequest)

	// 	assertStatusCode(t, response.Result().StatusCode, http.StatusOK)

	// 	wantEditNoteCalls := Notes{note}
	// 	assertAddNoteCallsEqual(t, notesStore.editNoteCalls, wantEditNoteCalls)
	// 	assertLoggingCalls(t, logger.infofCalls, []fmtCallf{
	// 		{format: "%s request to %s received", a: []any{"PUT", "/notes/1"}},
	// 	})
	// })
}

// WithUrlParam returns a pointer to a request object with the given URL params
// added to a new chi.Context object.
func WithUrlParam(r *http.Request, key, value string) *http.Request {
	chiCtx := chi.NewRouteContext()
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))
	chiCtx.URLParams.Add(key, value)
	return r
}

func encodeRequestBodyAddNote(t testing.TB, rb map[string]string) *bytes.Buffer {
	buf := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(buf).Encode(rb)
	assertNoError(t, err)
	return buf
}

func newGetNotesByUserIdRequest(t testing.TB, userID int) *http.Request {
	url := fmt.Sprintf("/users/%v/notes", userID)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("Could not build request newPostAddNoteRequest: %q", err)
	}
	return WithUrlParam(request, "userID", fmt.Sprintf("%v", userID))
}

func newGetAllNotesRequest(t testing.TB) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, "/notes", nil)
	if err != nil {
		t.Fatalf("Unable to build request newGetAllNotesRequest %q", err)
	}
	return req
}

func newDeleteRequest(t testing.TB, id int) *http.Request {
	url := fmt.Sprintf("/notes/%v", id)
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	assertNoError(t, err)
	request = WithUrlParam(request, "id", fmt.Sprintf("%d", id))
	return request
}

func newPostRequestWithNote(t testing.TB, note string) *http.Request {
	requestBody := map[string]string{"note": note}
	buf := encodeRequestBodyAddNote(t, requestBody)
	request, err := http.NewRequest(http.MethodPost, "", buf)
	assertNoError(t, err)
	return request
}

func newPostRequestFromBody(t testing.TB, requestBody string, url string) *http.Request {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(requestBody)
	assertNoError(t, err)
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	assertNoError(t, err)
	return req
}

// func newPutRequestWithNote(t testing.TB, note Note, url string) *http.Request {
// 	requestBody := map[string]Note{"note": note}
// 	buf := encodeRequestBodyAddNote(t, requestBody)
// 	request, err := http.NewRequest(http.MethodPut, url, buf)
// 	assertNoError(t, err)
// 	return request
// }

func newRequestWithBadIdParam(t testing.TB, badID string) *http.Request {
	request, err := http.NewRequest(http.MethodGet, "", nil)
	assertNoError(t, err)
	return WithUrlParam(request, "id", fmt.Sprintf("%v", badID))
}

func newInvalidBodyPostRequest(t testing.TB) *http.Request {
	requestBody := map[string]string{"wrong_key": "some text"}
	buf := encodeRequestBodyAddNote(t, requestBody)
	badRequest, err := http.NewRequest(http.MethodPost, "", buf)
	assertNoError(t, err)
	return badRequest
}

func assertLengthSlice[T any](t testing.TB, elements []T, want int) {
	t.Helper()
	if len(elements) != want {
		t.Errorf(`got = %v; want %v`, len(elements), want)
	}

}

func assertStatusCode(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf(`got = %v; want %v`, got, want)
	}
}

func assertSlicesHaveSameLength[T any](t testing.TB, got, want []T) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf(`len(got) = %v; len(want) %v`, len(got), len(want))
	}
}

func assertAddNoteCallsEqual(t testing.TB, gotCalls, wantCalls []AddNoteCall) {
	t.Helper()
	assertLengthSlice(t, gotCalls, len(wantCalls))
	for _, want := range wantCalls {
		found := false
		for _, got := range gotCalls {
			if reflect.DeepEqual(got, want) {
				found = true
			}
		}
		if !found {
			t.Errorf("want %v not found in gotCalls %v", want, gotCalls)
		}
	}
}

func assertLoggingCalls(t testing.TB, got, want []fmtCallf) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Error(`got unequal want`)
		logFormatfCalls(t, got)
		logFormatfCalls(t, want)
	}
}

func logFormatfCalls(t testing.TB, calls []fmtCallf) {
	t.Helper()
	for _, call := range calls {
		t.Log(call.String())
	}
}

func assertNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertRightErrorCall(t testing.TB, errorCall fmtCallf, wantFormat string, wantErr error) {
	t.Helper()
	gotFormat := errorCall.format
	if gotFormat != wantFormat {
		t.Errorf(`got = %v; want %v`, gotFormat, wantFormat)
	}
	if gotErr, ok := errorCall.a[0].(error); ok {
		if !errors.Is(gotErr, wantErr) {
			t.Errorf(`got = %v; want %v`, gotErr, wantErr)
		}
	} else {
		t.Errorf("Could not convert to error")
	}
}

func assertAllNotesGotCalled(t testing.TB, allNotesGotCalled bool) {
	t.Helper()
	if !allNotesGotCalled {
		t.Error("notesStore.AllNotes did not get called")
	}
}
func assertEqualIntSlice(t testing.TB, got, want []int) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf(`got = %v; want %v`, got, want)
	}
}
