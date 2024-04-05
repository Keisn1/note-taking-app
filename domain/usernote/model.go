package usernote

import (
	"github.com/Keisn1/note-taking-app/domain/entities"
	"github.com/google/uuid"
)

type UserNote struct {
	note *entities.Note
	user *entities.Person
}

func NewUserNote(title, content string, userID uuid.UUID) UserNote {
	return UserNote{
		note: &entities.Note{
			ID:      uuid.New(),
			Title:   entities.Title(title),
			Content: entities.Content(content),
		},
		user: &entities.Person{ID: userID},
	}
}

func (u UserNote) GetTitle() entities.Title {
	return u.note.Title
}

func (u UserNote) SetTitle(title string) {
	u.note.Title = entities.Title(title)
}

func (u UserNote) GetContent() entities.Content {
	return u.note.Content
}

func (u UserNote) SetContent(content string) {
	u.note.Content = entities.Content(content)
}

func (u UserNote) GetID() uuid.UUID {
	return u.note.ID
}

func (u UserNote) GetUserID() uuid.UUID {
	return u.user.ID
}
