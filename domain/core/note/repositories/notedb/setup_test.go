package notedb_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/Keisn1/note-taking-app/domain/core/note"
	"github.com/google/uuid"
)

func run(m *testing.M) int {
	var (
		dropDB   = fmt.Sprintf(`DROP DATABASE IF EXISTS %s;`, testDBName)
		createDB = fmt.Sprintf(`CREATE DATABASE %s;`, testDBName)
	)

	dsn := fmt.Sprintf("host=localhost port=5432 user=%s password=%s sslmode=disable", testUser, testPassword)
	postgresDB, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(err)
	}
	defer postgresDB.Close()

	_, err = postgresDB.Exec(dropDB)
	if err != nil {
		panic(err)
	}

	_, err = postgresDB.Exec(createDB)
	if err != nil {
		panic(err)
	}

	defer func() {
		_, err = postgresDB.Exec(dropDB)
		if err != nil {
			panic(fmt.Errorf("postgresDB.Exec() err = %s", err))
		}
	}()

	return m.Run()
}

func SetupNotesTable(t *testing.T, notes []note.Note) (*sql.DB, func()) {
	var (
		createNoteTable = `CREATE TABLE notes(
							id UUID PRIMARY KEY,
							title TEXT,
							content TEXT,
							user_id UUID NOT NULL)`
		dropNotesTable = `DROP TABLE notes`
	)

	dsn := fmt.Sprintf("host=localhost port=5432 user=%s password=%s sslmode=disable dbname=%s ", testUser, testPassword, testDBName)
	testDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}

	_, err = testDB.Exec(createNoteTable)
	if err != nil {
		t.Fatal(err)
	}

	insertRow := `INSERT INTO notes (id, title, content, user_id) VALUES ($1, $2, $3, $4)`
	for _, n := range notes {
		_, err = testDB.Exec(
			insertRow,
			n.ID,
			n.Title.String(),
			n.Content.String(),
			n.UserID,
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	deleteTable := func() {
		_, err := testDB.Exec(dropNotesTable)
		if err != nil {
			t.Fatal(err)
		}
		testDB.Close()
	}

	return testDB, deleteTable
}

func fixtureNotes() []note.Note {
	return []note.Note{
		{ID: uuid.UUID{1}, Title: note.NewTitle("robs 1st note"), Content: note.NewContent("robs 1st note content"), UserID: uuid.UUID{1}},
		{ID: uuid.UUID{2}, Title: note.NewTitle("robs 2nd note"), Content: note.NewContent("robs 2nd note content"), UserID: uuid.UUID{1}},
		{ID: uuid.UUID{3}, Title: note.NewTitle("annas 1st note"), Content: note.NewContent("annas 1st note content"), UserID: uuid.UUID{2}},
		{ID: uuid.UUID{4}, Title: note.NewTitle("annas 2nd note"), Content: note.NewContent("annas 2nd note content"), UserID: uuid.UUID{2}},
	}
}
