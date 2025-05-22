package models

type Character struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Strength        int    `json:"strength"`
	Agility         int    `json:"agility"`
	Stamina         int    `json:"stamina"`
	Level           int    `json:"level"`
	XP              int    `json:"xp"`
	AvailablePoints int    `json:"available_points"`
}

// Player represents a player in a lobby, linking to a Character
type Player struct {
	CharacterID string `json:"character_id"`
	Name        string `json:"name"` // Denormalized from Character for convenience
}

// Lobby represents a game lobby where players can gather before a match
type Lobby struct {
	LobbyID    string   `json:"lobby_id"`
	Players    []Player `json:"players"`
	MaxPlayers int      `json:"max_players"`
}

// BodyPart represents a part of the body that can be targeted or blocked.
type BodyPart string

const (
	Head  BodyPart = "Head"
	Chest BodyPart = "Chest"
	Groin BodyPart = "Groin"
	Legs  BodyPart = "Legs"
)

// Action represents a player's chosen action for a turn in a fight.
type Action struct {
	AttackerID string   `json:"attacker_id"` // ID of the character performing the action
	TargetID   string   `json:"target_id"`   // ID of the character being targeted
	AttackArea BodyPart `json:"attack_area"` // Area the attacker targets
	BlockAreas []BodyPart `json:"block_areas"` // Areas the attacker chooses to block (max 2)
}

// FightPlayerData represents a player's state within a fight.
type FightPlayerData struct {
	Character       Character `json:"character"`         // Embedded character data
	CurrentHealth   int       `json:"current_health"`    // Current health in the fight
	ChosenAttackArea BodyPart `json:"chosen_attack_area"` // Set per turn
	ChosenBlockAreas []BodyPart `json:"chosen_block_areas"` // Set per turn
}

// Fight represents the state of a combat encounter.
type Fight struct {
	FightID             string          `json:"fight_id"`
	Player1             FightPlayerData `json:"player1"`
	Player2             FightPlayerData `json:"player2"` // Could be CPU or another player
	Turn                int             `json:"turn"`
	IsOver              bool            `json:"is_over"`
	WinnerID            string          `json:"winner_id,omitempty"` // Optional: ID of the winning character
	CurrentTurnPlayerID string          `json:"current_turn_player_id"`
}
