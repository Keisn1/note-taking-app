package note_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Keisn1/note-taking-app/domain/core/note"
	"github.com/Keisn1/note-taking-app/domain/core/note/repositories/memory"
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
		noteID := robsNote.GetID()

		err := notesS.Delete(noteID)
		assert.NoError(t, err)

		_, err = notesS.GetNoteByID(noteID)
		assert.ErrorContains(t, err, fmt.Errorf("getNoteByID: [%s]", noteID).Error())
	})
}

func TestNoteService_Create(t *testing.T) {
	t.Run("Throws error if userID not present", func(t *testing.T) {
		notesS := Setup(t, fixtureNotes())
		userID := uuid.New()

		newNote := note.NewUpdateNote("invalid title", "", userID)
		_, err := notesS.Create(newNote)
		assert.Error(t, err)
	})

	t.Run("Throws error if repo throws error (given repo.Create is called)", func(t *testing.T) {
		errorRepo := ErrorNoteRepo{}
		notesS := note.NewNotesService(errorRepo, user.UserSvc{})

		userID := uuid.New()
		newNote := note.NewUpdateNote("invalid title", "", userID)
		_, err := notesS.Create(newNote)
		assert.Error(t, err)
	})

	t.Run("I can create a new note", func(t *testing.T) {
		notesS := Setup(t, fixtureNotes())

		userID := uuid.UUID{1}
		newNote := note.NewUpdateNote("new note title", "new note content", userID)

		got, err := notesS.Create(newNote)
		assert.NoError(t, err)
		assert.NotEqual(t, got.GetID(), uuid.UUID{})
		assert.Equal(t, "new note title", got.GetTitle().String())
		assert.Equal(t, "new note content", got.GetContent().String())
		assert.Equal(t, userID, got.GetUserID())

		noteID := got.GetID()
		want := got
		got, err = notesS.GetNoteByID(noteID)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

func TestNoteService_Update(t *testing.T) {
	// t.Run("Given a note NOT present in the system and a note containing updates for this note, it throws an error", func(t *testing.T) {
	// 	notesR, err := memory.NewNotesRepo(fixtureNotes())
	// 	assert.NoError(t, err)
	// 	notesS := note.NewNotesService(notesR)

	// 	_, err = notesS.Update(note.Note{}, note.UpdateNote{})
	// 	assert.ErrorContains(t, err, "update: ")
	// })

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
				name:       "New Title, 0length content, update of both: 'new title' and ''",
				currNote:   note.NewNote(uuid.UUID{1}, "robs 1st note", "robs 1st note content", uuid.UUID{1}),
				updateNote: note.NewUpdateNote("new title", "", uuid.UUID{1}),
				want:       note.NewNote(uuid.UUID{1}, "new title", "", uuid.UUID{1}),
			},
			{
				name:       "New Title, empty content, will update only title: 'new title'",
				currNote:   note.NewNote(uuid.UUID{2}, "robs 2nd note", "robs 2nd note content", uuid.UUID{1}),
				updateNote: note.UpdateNote{Title: note.NewTitle("new title"), Content: note.Content{}, UserID: uuid.UUID{1}},
				want:       note.NewNote(uuid.UUID{2}, "new title", "robs 2nd note content", uuid.UUID{1}),
			},
			{
				name:       "0length title, New content, update of both: '' and 'new content'",
				currNote:   note.NewNote(uuid.UUID{3}, "annas 1st note", "annas 1st note content", uuid.UUID{2}),
				updateNote: note.NewUpdateNote("", "new content", uuid.UUID{2}),
				want:       note.NewNote(uuid.UUID{3}, "", "new content", uuid.UUID{2}),
			},
			{
				name:       "empty title, empty content, therefore no update at all",
				currNote:   note.NewNote(uuid.UUID{4}, "annas 2nd note", "annas 2nd note content", uuid.UUID{2}),
				updateNote: note.UpdateNote{Title: note.Title{}, Content: note.Content{}, UserID: uuid.UUID{2}},
				want:       note.NewNote(uuid.UUID{4}, "annas 2nd note", "annas 2nd note content", uuid.UUID{2}),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := notesS.Update(tc.currNote, tc.updateNote)
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got) // assert that the right note was sent back

				got, err = notesS.GetNoteByID(tc.currNote.GetID())
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got) // asssert that the note can actually be retrieved
			})
		}
	})
}

func TestNoteService_GetNoteByID(t *testing.T) {
	t.Run("GetNoteByID return error on missing note", func(t *testing.T) {
		notesS := Setup(t, fixtureNotes())
		noteID := uuid.New()
		_, err := notesS.GetNoteByID(noteID)
		assert.ErrorContains(t, err, fmt.Errorf("getNoteByID: [%s]", noteID).Error())
	})

	t.Run("I can get a note by its ID", func(t *testing.T) {
		notesS := Setup(t, fixtureNotes())
		type testCase struct {
			noteID uuid.UUID
			want   note.Note
		}

		testCases := []testCase{
			{noteID: uuid.UUID{1}, want: note.NewNote(uuid.UUID{1}, "robs 1st note", "robs 1st note content", uuid.UUID{1})},
			{noteID: uuid.UUID{3}, want: note.NewNote(uuid.UUID{3}, "annas 1st note", "annas 1st note content", uuid.UUID{2})},
		}

		for _, tc := range testCases {
			got, err := notesS.GetNoteByID(tc.noteID)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		}
	})

	t.Run("GetNoteByUserID return errors on missing user", func(t *testing.T) {
		notesS := Setup(t, fixtureNotes())
		userID := uuid.New()
		_, err := notesS.GetNotesByUserID(userID)
		assert.ErrorContains(t, err, fmt.Errorf("getNoteByUserID: [%s]", userID).Error())
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
					note.NewNote(uuid.UUID{}, "robs 1st note", "robs 1st note content", uuid.UUID{1}),
					note.NewNote(uuid.UUID{}, "robs 2nd note", "robs 2nd note content", uuid.UUID{1}),
				},
			},
			{
				userID: uuid.UUID{2},
				want: []note.Note{
					note.NewNote(uuid.UUID{}, "annas 1st note", "annas 1st note content", uuid.UUID{2}),
					note.NewNote(uuid.UUID{}, "annas 2nd note", "annas 2nd note content", uuid.UUID{2}),
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

func fixtureNotes() []note.Note {
	return []note.Note{
		note.NewNote(uuid.UUID{1}, "robs 1st note", "robs 1st note content", uuid.UUID{1}),
		note.NewNote(uuid.UUID{2}, "robs 2nd note", "robs 2nd note content", uuid.UUID{1}),
		note.NewNote(uuid.UUID{3}, "annas 1st note", "annas 1st note content", uuid.UUID{2}),
		note.NewNote(uuid.UUID{4}, "annas 2nd note", "annas 2nd note content", uuid.UUID{2}),
	}
}

func Setup(t *testing.T, notes []note.Note) note.NotesService {
	t.Helper()
	notesR, err := memory.NewNotesRepo(notes)
	assert.NoError(t, err)

	userSvc := StubUserService{ids: make(map[uuid.UUID]struct{})}
	for _, n := range notes {
		userSvc.ids[n.UserID] = struct{}{}
	}

	return note.NewNotesService(notesR, userSvc)
}

type StubUserService struct {
	ids map[uuid.UUID]struct{}
}

func (sus StubUserService) QueryByID(userID uuid.UUID) (user.User, error) {
	if _, ok := sus.ids[userID]; !ok {
		return user.User{}, errors.New("User not found")
	}
	return user.User{ID: userID}, nil
}
