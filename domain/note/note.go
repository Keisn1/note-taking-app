package note

import (
	"fmt"
	"github.com/google/uuid"
)

type Note struct {
	NoteID  uuid.UUID
	Title   string
	Content string
	UserID  uuid.UUID
}

type NoteRepo struct {
	notes map[uuid.UUID]Note
}

func NewNotesRepo(notes []Note) (NoteRepo, error) {
	var nR NoteRepo
	if err := noDuplicate(notes); err != nil {
		return NoteRepo{}, fmt.Errorf("newNotesRepo: %w", err)
	}

	nR.notes = make(map[uuid.UUID]Note)
	for _, n := range notes {
		nR.notes[n.NoteID] = n
	}
	return nR, nil
}

func (nR NoteRepo) Update(noteID uuid.UUID, newTitle string) {
	n := nR.notes[noteID]
	n.Title = newTitle
	nR.notes[noteID] = n
}

func (nR NoteRepo) GetNoteByID(noteID uuid.UUID) Note {
	for _, n := range nR.notes {
		if n.NoteID == noteID {
			return n
		}
	}
	return Note{}
}

func (nR NoteRepo) GetNotesByUserID(userID uuid.UUID) []Note {
	var ret []Note
	for _, n := range nR.notes {
		if n.UserID == userID {
			n.NoteID = uuid.UUID{0}
			ret = append(ret, n)
		}
	}
	return ret
}

func noDuplicate(notes []Note) error {
	noteIDSet := make(map[uuid.UUID]struct{})
	for _, n := range notes {
		if _, ok := noteIDSet[n.NoteID]; ok {
			return fmt.Errorf("duplicate noteID [%s]", n.NoteID)
		}
		noteIDSet[n.NoteID] = struct{}{}
	}
	return nil
}
