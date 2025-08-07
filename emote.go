package emote

import "github.com/google/uuid"

type Emote struct {
	UUID     uuid.UUID
	Name     string
	Rarity   string
	Keywords []string
}
