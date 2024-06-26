package note_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Keisn1/note-taking-app/domain/core/note"
	"github.com/Keisn1/note-taking-app/domain/core/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNoteService_Delete(t *testing.T) {
	t.Run("Try to delete a non present note gives an error", func(t *testing.T) {
		notesS := Setup(t, fixtureNotes())
		noteID := uuid.UUID{}

		err := notesS.Delete(noteID)
		assert.ErrorContains(t, err, fmt.Errorf("delete: [%s]", noteID).Error())
	})

	t.Run("I can delete a note by its ID", func(t *testing.T) {
		notesS := Setup(t, fixtureNotes())
		robsNote := fixtureNotes()[0]
		noteID := robsNote.ID

		err := notesS.Delete(noteID)
		assert.NoError(t, err)

		_, err = notesS.QueryByID(context.Background(), noteID)
		assert.ErrorContains(t, err, fmt.Errorf("getNoteByID: [%s]", noteID).Error())
	})
}

func TestNoteService_Create(t *testing.T) {
	t.Run("Throws error if userID not present", func(t *testing.T) {
		notesS := Setup(t, fixtureNotes())
		userID := uuid.New()

		newNote := note.UpdateNote{Title: note.NewTitle(""), Content: note.NewContent(""), UserID: userID}
		_, err := notesS.Create(context.Background(), newNote)
		assert.Error(t, err)
	})

	t.Run("Throws error if repo throws error (given repo.Create is called)", func(t *testing.T) {
		errorRepo := ErrorNoteRepo{}
		notesS := note.NewNotesService(errorRepo, user.Svc{})

		userID := uuid.New()
		newNote := note.UpdateNote{Title: note.NewTitle(""), Content: note.NewContent(""), UserID: userID}
		_, err := notesS.Create(context.Background(), newNote)
		assert.Error(t, err)
	})

	t.Run("I can create a new note", func(t *testing.T) {
		notesS := Setup(t, fixtureNotes())

		userID := uuid.UUID{1}

		newNote := note.UpdateNote{Title: note.NewTitle("new note title"), Content: note.NewContent("new note content"), UserID: userID}

		got, err := notesS.Create(context.Background(), newNote)
		assert.NoError(t, err)
		assert.NotEqual(t, got.ID, uuid.UUID{})
		assert.Equal(t, "new note title", got.Title.String())
		assert.Equal(t, "new note content", got.Content.String())
		assert.Equal(t, userID, got.UserID)

		noteID := got.ID
		want := got

		got, err = notesS.QueryByID(context.Background(), noteID)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

func TestNoteService_Update(t *testing.T) {
	t.Run("Given a note present in the system and a note containing updates for this note, I can update the present note inside the system", func(t *testing.T) {
		notesS := Setup(t, fixtureNotes())

		type testCase struct {
			name       string
			currNote   note.Note
			updateNote note.UpdateNote
			want       note.Note
		}

		testCases := []testCase{
			{
				name: "New Title, 0length content, update of both: 'new title' and ''",
				currNote: note.Note{
					ID: uuid.UUID{1}, Title: note.NewTitle("robs 1st note"), Content: note.NewContent("robs 1st note content"), UserID: uuid.UUID{1},
				},
				updateNote: note.UpdateNote{Title: note.NewTitle("new title"), Content: note.NewContent(""), UserID: uuid.UUID{1}},
				want: note.Note{
					ID: uuid.UUID{1}, Title: note.NewTitle("new title"), Content: note.NewContent(""), UserID: uuid.UUID{1},
				},
			},
			{
				name: "New Title, empty content, will update only title: 'new title'",
				currNote: note.Note{
					ID: uuid.UUID{1}, Title: note.NewTitle("robs 1st note"), Content: note.NewContent("robs 1st note content"), UserID: uuid.UUID{1},
				},
				updateNote: note.UpdateNote{Title: note.NewTitle("new title"), Content: note.Content{}, UserID: uuid.UUID{1}},
				want: note.Note{
					ID: uuid.UUID{1}, Title: note.NewTitle("new title"), Content: note.NewContent("robs 1st note content"), UserID: uuid.UUID{1},
				},
			},
			{
				name: "0length title, New content, update of both: '' and 'new content'",
				currNote: note.Note{
					ID: uuid.UUID{1}, Title: note.NewTitle("robs 1st note"), Content: note.NewContent("robs 1st note content"), UserID: uuid.UUID{1},
				},
				updateNote: note.UpdateNote{Title: note.NewTitle(""), Content: note.NewContent("new content"), UserID: uuid.UUID{2}},
				want: note.Note{
					ID: uuid.UUID{1}, Title: note.NewTitle(""), Content: note.NewContent("new content"), UserID: uuid.UUID{1},
				},
			},
			{
				name: "empty title, empty content, therefore no update at all",
				currNote: note.Note{
					ID: uuid.UUID{1}, Title: note.NewTitle("robs 1st note"), Content: note.NewContent("robs 1st note content"), UserID: uuid.UUID{1},
				},
				updateNote: note.UpdateNote{Title: note.Title{}, Content: note.Content{}, UserID: uuid.UUID{2}},
				want: note.Note{
					ID: uuid.UUID{1}, Title: note.NewTitle("robs 1st note"), Content: note.NewContent("robs 1st note content"), UserID: uuid.UUID{1},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := notesS.Update(tc.currNote, tc.updateNote)
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got) // assert that the right note was sent back

				got, err = notesS.QueryByID(context.Background(), tc.currNote.ID)
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got) // asssert that the note can actually be retrieved
			})
		}
	})
}

func TestNoteService_QueryByID(t *testing.T) {
	t.Run("GetNoteByID return error on missing note", func(t *testing.T) {
		notesS := Setup(t, fixtureNotes())
		noteID := uuid.New()
		_, err := notesS.QueryByID(context.Background(), noteID)
		assert.ErrorContains(t, err, fmt.Errorf("getNoteByID: [%s]", noteID).Error())
	})

	t.Run("GetNoteByUserID return errors on missing user", func(t *testing.T) {
		notesS := Setup(t, fixtureNotes())
		userID := uuid.New()
		_, err := notesS.GetNotesByUserID(userID)
		assert.ErrorContains(t, err, fmt.Errorf("getNoteByUserID: [%s]", userID).Error())
	})

	t.Run("I can get a note by its ID", func(t *testing.T) {
		notesS := Setup(t, fixtureNotes())
		type testCase struct {
			noteID uuid.UUID
			want   note.Note
		}

		testCases := []testCase{
			{
				noteID: uuid.UUID{1},
				want: note.Note{
					ID: uuid.UUID{1}, Title: note.NewTitle("robs 1st note"), Content: note.NewContent("robs 1st note content"), UserID: uuid.UUID{1},
				},
			},
			{
				noteID: uuid.UUID{3},
				want: note.Note{
					ID: uuid.UUID{3}, Title: note.NewTitle("annas 1st note"), Content: note.NewContent("annas 1st note content"), UserID: uuid.UUID{2},
				},
			},
		}

		for _, tc := range testCases {
			got, err := notesS.QueryByID(context.Background(), tc.noteID)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		}
	})
}

func TestNoteService_GetNotesByUserID(t *testing.T) {
	t.Run("I can get all notes of a User by the userID", func(t *testing.T) {
		notesS := Setup(t, fixtureNotes())
		type testCase struct {
			userID uuid.UUID
			want   []note.Note
		}

		testCases := []testCase{
			{
				userID: uuid.UUID{1},
				want: []note.Note{
					{ID: uuid.UUID{1}, Title: note.NewTitle("robs 1st note"), Content: note.NewContent("robs 1st note content"), UserID: uuid.UUID{1}},
					{ID: uuid.UUID{2}, Title: note.NewTitle("robs 2nd note"), Content: note.NewContent("robs 2nd note content"), UserID: uuid.UUID{1}},
				},
			},
			{
				userID: uuid.UUID{2},
				want: []note.Note{
					{ID: uuid.UUID{3}, Title: note.NewTitle("annas 1st note"), Content: note.NewContent("annas 1st note content"), UserID: uuid.UUID{2}},
					{ID: uuid.UUID{4}, Title: note.NewTitle("annas 2nd note"), Content: note.NewContent("annas 2nd note content"), UserID: uuid.UUID{2}},
				},
			},
		}

		for _, tc := range testCases {
			got, err := notesS.GetNotesByUserID(tc.userID)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tc.want, got)
		}
	})
}
