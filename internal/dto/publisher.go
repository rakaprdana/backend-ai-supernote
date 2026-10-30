package dto

import "github.com/google/uuid"

type PublisherEmbedNoteMessage struct {
	NoteId uuid.UUID `json:"note_id"`
}
