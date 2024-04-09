package postgres_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"database/sql"

	"github.com/Keisn1/note-taking-app/domain/note"
	"github.com/Keisn1/note-taking-app/domain/note/repositories/postgres"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
)

const (
	testDBName   = "test_note_taking_app"
	testUser     = "postgres"
	testPassword = "password"
)

func TestMain(m *testing.M) {
	exitCode := run(m)
	os.Exit(exitCode)
}

func TestNotesRepo_Update(t *testing.T) {
	testDB, deleteTable := SetupNotesTable(t, fixtureNotes())
	defer testDB.Close()
	defer deleteTable()

	t.Run("Able to delete a note", func(t *testing.T) {

	})
}

func TestNotesRepo_Edit(t *testing.T) {
	testDB, deleteTable := SetupNotesTable(t, fixtureNotes())
	defer testDB.Close()
	defer deleteTable()

	t.Run("Given a note present in the system, I can update its title and its content", func(t *testing.T) {
		nR := postgres.NewNotesRepo(testDB)

		type testCase struct {
			name       string
			currNote   note.Note
			updateNote note.Note
			want       note.Note
		}

		testCases := []testCase{
			{
				name:       "New Title, new 0 character content, update of both (title and content) expected",
				currNote:   note.MakeNote(uuid.UUID{1}, note.NewTitle("robs 1st note"), note.NewContent("robs 1st note content"), uuid.UUID{1}),
				updateNote: note.MakeNote(uuid.UUID{1}, note.NewTitle("New title"), note.NewContent(""), uuid.UUID{1}),
				want:       note.MakeNote(uuid.UUID{1}, note.NewTitle("New title"), note.NewContent(""), uuid.UUID{1}),
			},
			{
				name:       "New Title, empty content, therefore no update",
				currNote:   note.MakeNote(uuid.UUID{2}, note.NewTitle("robs 2nd note"), note.NewContent("robs 2nd note content"), uuid.UUID{1}),
				updateNote: note.MakeNote(uuid.UUID{2}, note.NewTitle("New title"), note.Content{}, uuid.UUID{1}),
				want:       note.MakeNote(uuid.UUID{2}, note.NewTitle("New title"), note.NewContent("robs 2nd note content"), uuid.UUID{1}),
			},
			{
				name:       "New 0 character title, new content, update of both (title and content) expected",
				currNote:   note.MakeNote(uuid.UUID{3}, note.NewTitle("annas 1st note"), note.NewContent("annas 1st note content"), uuid.UUID{2}),
				updateNote: note.MakeNote(uuid.UUID{3}, note.NewTitle(""), note.NewContent("New content"), uuid.UUID{2}),
				want:       note.MakeNote(uuid.UUID{3}, note.NewTitle(""), note.NewContent("New content"), uuid.UUID{2}),
			},
			{
				name:       "empty title, empty content, therefore no update",
				currNote:   note.MakeNote(uuid.UUID{4}, note.NewTitle("annas 2nd note"), note.NewContent("annas 2nd note content"), uuid.UUID{2}),
				updateNote: note.MakeNote(uuid.UUID{4}, note.Title{}, note.Content{}, uuid.UUID{2}),
				want:       note.MakeNote(uuid.UUID{4}, note.NewTitle("annas 2nd note"), note.NewContent("annas 2nd note content"), uuid.UUID{2}),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				n, err := nR.Update(tc.currNote, tc.updateNote)
				assert.NoError(t, err)
				// assert.Equal(t, tc.want, n) // assert that the right note was sent back

				// got, err := notesS.GetNoteByID(tc.currNote.GetID())
				// assert.NoError(t, err)
				// assert.Equal(t, tc.want, got) // asssert that the note can actually be retrieved
			})
		}
	})
}

func TestNotesRepo_Delete(t *testing.T) {
	testDB, deleteTable := SetupNotesTable(t, fixtureNotes())
	defer testDB.Close()
	defer deleteTable()

	t.Run("Able to delete a note", func(t *testing.T) {
		nR := postgres.NewNotesRepo(testDB)

		nN := note.MakeNewNote(note.NewTitle("title"), note.NewContent("content"), uuid.UUID{1})
		n := note.MakeNoteFromNewNote(nN)
		nR.Create(n)

		err := nR.Delete(n.GetID())
		assert.NoError(t, err)
	})
	t.Run("Delete non-present note throws error ", func(t *testing.T) {
		nR := postgres.NewNotesRepo(testDB)

		noteID := uuid.New()
		err := nR.Delete(noteID)
		assert.ErrorContains(t, err, fmt.Sprintf("delete: note not present [%s]", noteID))
	})
}

func TestNotesRepo_Create(t *testing.T) {
	t.Run("Add a note", func(t *testing.T) {
		testDB, deleteTable := SetupNotesTable(t, []note.Note{})
		defer testDB.Close()
		defer deleteTable()
		nR := postgres.NewNotesRepo(testDB)

		n := note.MakeNote(uuid.UUID{1}, note.NewTitle("robs 1st note"), note.NewContent("robs 1st note content"), uuid.UUID{1})
		err := nR.Create(n)
		assert.NoError(t, err)

		got, err := nR.GetNoteByID(n.GetID())
		assert.NoError(t, err)
		assert.Equal(t, got, n)
	})

	t.Run("Throws error if note to be created is already present", func(t *testing.T) {
		testDB, deleteTable := SetupNotesTable(t, fixtureNotes())
		defer testDB.Close()
		defer deleteTable()
		nR := postgres.NewNotesRepo(testDB)

		n := note.MakeNote(uuid.UUID{1}, note.NewTitle("robs 1st note"), note.NewContent("robs 1st note content"), uuid.UUID{1})
		err := nR.Create(n)
		assert.Error(t, err)
		assert.ErrorContains(t, err, fmt.Sprintf("create: [%s]", n.GetID()))
	})
}

func TestNotesRepo_GetNoteByID(t *testing.T) {
	testDB, deleteTable := SetupNotesTable(t, fixtureNotes())
	defer testDB.Close()
	defer deleteTable()

	t.Run("Happy: get note by id", func(t *testing.T) {
		nR := postgres.NewNotesRepo(testDB)

		type testCase struct {
			noteID uuid.UUID
			want   note.Note
		}

		testCases := []testCase{
			{noteID: uuid.UUID{1}, want: note.MakeNote(uuid.UUID{1}, note.NewTitle("robs 1st note"), note.NewContent("robs 1st note content"), uuid.UUID{1})},
			{noteID: uuid.UUID{3}, want: note.MakeNote(uuid.UUID{3}, note.NewTitle("annas 1st note"), note.NewContent("annas 1st note content"), uuid.UUID{2})},
		}

		for _, tc := range testCases {
			got, err := nR.GetNoteByID(tc.noteID)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		}
	})

	t.Run("Note not found - sql.ErrNoRows", func(t *testing.T) {
		nR := postgres.NewNotesRepo(testDB)
		noteID := uuid.UUID{}

		_, err := nR.GetNoteByID(noteID)
		assert.EqualError(t, err, fmt.Errorf("getNoteByID: not found [%s]: %w", noteID, sql.ErrNoRows).Error())
	})
}

func TestNotesRepo_GetNotesByUserID(t *testing.T) {
	testDB, deleteTable := SetupNotesTable(t, fixtureNotes())
	defer testDB.Close()
	defer deleteTable()

	t.Run("Get notes by userID", func(t *testing.T) {
		nR := postgres.NewNotesRepo(testDB)
		type testCase struct {
			userID uuid.UUID
			want   []note.Note
		}

		testCases := []testCase{
			{
				userID: uuid.UUID{1},
				want: []note.Note{
					note.MakeNote(uuid.UUID{1}, note.NewTitle("robs 1st note"), note.NewContent("robs 1st note content"), uuid.UUID{1}),
					note.MakeNote(uuid.UUID{2}, note.NewTitle("robs 2nd note"), note.NewContent("robs 2nd note content"), uuid.UUID{1}),
				},
			},
			{
				userID: uuid.UUID{2},
				want: []note.Note{
					note.MakeNote(uuid.UUID{3}, note.NewTitle("annas 1st note"), note.NewContent("annas 1st note content"), uuid.UUID{2}),
					note.MakeNote(uuid.UUID{4}, note.NewTitle("annas 2nd note"), note.NewContent("annas 2nd note content"), uuid.UUID{2}),
				},
			},
		}

		for _, tc := range testCases {
			got, err := nR.GetNotesByUserID(tc.userID)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tc.want, got)
		}
	})
	t.Run("Returns error on missing user", func(t *testing.T) {
		nR := postgres.NewNotesRepo(testDB)

		userID := uuid.UUID{}
		wantErrMsg := fmt.Sprintf("getNotesByUserID: not found [%s]", userID)
		_, err := nR.GetNotesByUserID(userID)
		assert.ErrorContains(t, err, wantErrMsg)
	})

	t.Run("Fowards error on database error", func(t *testing.T) {
		sdb := &SQLDB{}
		nR := postgres.NewNotesRepo(sdb)

		userID := uuid.UUID{}
		wantErr := fmt.Errorf("getNotesByUserID: [%s]: %w", userID, errors.New("DBError"))
		_, err := nR.GetNotesByUserID(userID)
		assert.EqualError(t, err, wantErr.Error())
	})

}