package api

import (
	"encoding/json"
	"fmt"
	"gladiator-backend/models"
	"net/http"
	"strings" // Added import
	"sync"

	"github.com/google/uuid"
)

var (
	characters   = make(map[string]models.Character)
	lobbies      = make(map[string]models.Lobby)
	activeFights = make(map[string]*models.Fight) // New map for active fights
	mutex        = &sync.Mutex{}                 // Shared mutex for characters, lobbies, and fights
)

// CreateCharacter handles POST requests to /api/characters
// It creates a new character with default stats.
func CreateCharacter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Character name cannot be empty", http.StatusBadRequest)
		return
	}

	newCharacter := models.Character{
		ID:              uuid.New().String(),
		Name:            req.Name,
		Strength:        10,
		Agility:         10,
		Stamina:         10,
		Level:           1,
		XP:              0,
		AvailablePoints: 5,
	}

	mutex.Lock()
	characters[newCharacter.ID] = newCharacter
	mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newCharacter); err != nil {
		// If encoding fails, log it and send a generic error.
		// This is important because headers might have already been sent.
		fmt.Printf("Error encoding character: %v\n", err)
		http.Error(w, "Failed to encode character", http.StatusInternalServerError)
	}
}

// GetCharacter handles GET requests to /api/characters/{id}
// It retrieves a character by its ID.
func GetCharacter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path, e.g., /api/characters/some-uuid
	// This requires a router that can handle path parameters.
	// For now, we'll assume the ID is passed as a query parameter for simplicity with net/http default mux
	// e.g. /api/character?id=some-uuid
	// A better approach would be to use a router like gorilla/mux, or parse the path.
	// For now, let's try to parse from path: /api/characters/{id}
	// Path will be like "/api/characters/uuid_string"
	// We need to extract "uuid_string"

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 || pathParts[3] == "" { // Expecting /api/characters/{id}
		http.Error(w, "Character ID is missing in URL path", http.StatusBadRequest)
		return
	}
	id := pathParts[3]


	mutex.Lock()
	character, ok := characters[id]
	mutex.Unlock()

	if !ok {
		http.Error(w, "Character not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(character); err != nil {
		fmt.Printf("Error encoding character: %v\n", err)
		http.Error(w, "Failed to encode character", http.StatusInternalServerError)
	}
}

// UpdateCharacterAttributes handles PUT requests to /api/characters/{id}/attributes
// It updates a character's attributes if available points are sufficient.
func UpdateCharacterAttributes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	// Expecting /api/characters/{id}/attributes
	if len(pathParts) < 5 || pathParts[3] == "" {
		http.Error(w, "Character ID is missing or path is malformed", http.StatusBadRequest)
		return
	}
	id := pathParts[3]

	var req struct {
		Strength int `json:"strength"`
		Agility  int `json:"agility"`
		Stamina  int `json:"stamina"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	pointsToSpend := req.Strength + req.Agility + req.Stamina
	if pointsToSpend <= 0 {
		http.Error(w, "No attribute points provided or invalid values", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock() // Ensure mutex is unlocked even if errors occur

	character, ok := characters[id]
	if !ok {
		http.Error(w, "Character not found", http.StatusNotFound)
		return
	}

	if character.AvailablePoints < pointsToSpend {
		http.Error(w, fmt.Sprintf("Not enough available points. Available: %d, Requested: %d", character.AvailablePoints, pointsToSpend), http.StatusBadRequest)
		return
	}

	character.Strength += req.Strength
	character.Agility += req.Agility
	character.Stamina += req.Stamina
	character.AvailablePoints -= pointsToSpend
	characters[id] = character

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(character); err != nil {
		fmt.Printf("Error encoding updated character: %v\n", err)
		http.Error(w, "Failed to encode updated character", http.StatusInternalServerError)
	}
}

// CreateLobby handles POST requests to /api/lobbies
// It creates a new lobby.
func CreateLobby(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	newLobby := models.Lobby{
		LobbyID:    uuid.New().String(),
		Players:    []models.Player{},
		MaxPlayers: 2, // Default to 2 players
	}

	mutex.Lock()
	lobbies[newLobby.LobbyID] = newLobby
	mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// Respond with { "lobbyId": "your-new-lobby-id" }
	if err := json.NewEncoder(w).Encode(map[string]string{"lobbyId": newLobby.LobbyID}); err != nil {
		fmt.Printf("Error encoding lobby ID: %v\n", err)
		http.Error(w, "Failed to encode lobby ID", http.StatusInternalServerError)
	}
}

// JoinLobby handles POST requests to /api/lobbies/{lobbyId}/join
// It adds a character to a lobby.
func JoinLobby(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	// Expecting /api/lobbies/{lobbyId}/join
	if len(pathParts) < 5 || pathParts[4] != "join" {
		http.Error(w, "Lobby ID is missing or path is malformed", http.StatusBadRequest)
		return
	}
	lobbyID := pathParts[3]

	var req struct {
		CharacterID string `json:"characterId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CharacterID == "" {
		http.Error(w, "Character ID cannot be empty", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	lobby, ok := lobbies[lobbyID]
	if !ok {
		http.Error(w, "Lobby not found", http.StatusNotFound)
		return
	}

	character, charExists := characters[req.CharacterID]
	if !charExists {
		http.Error(w, "Character not found", http.StatusNotFound) // Or BadRequest, depending on desired semantics
		return
	}

	if len(lobby.Players) >= lobby.MaxPlayers {
		http.Error(w, "Lobby is full", http.StatusConflict) // 409 Conflict is suitable here
		return
	}

	// Check if player (character) is already in the lobby
	for _, p := range lobby.Players {
		if p.CharacterID == req.CharacterID {
			http.Error(w, "Character already in lobby", http.StatusBadRequest)
			return
		}
	}

	newPlayer := models.Player{
		CharacterID: character.ID,
		Name:        character.Name, // Denormalized from character
	}

	lobby.Players = append(lobby.Players, newPlayer)
	lobbies[lobbyID] = lobby // Update the lobby in the map

	w.Header().Set("Content-Type", "application/json")
	// Respond with the updated lobby object as per README suggestion (lobbyId and players)
	// The README suggests: `{ "lobbyId": 1, "players": [{ "id": 1, "name": "Gladiator's Name", ... }] }`
	// Our models.Player has CharacterID and Name.
	response := struct {
		LobbyID string          `json:"lobbyId"`
		Players []models.Player `json:"players"`
	}{
		LobbyID: lobby.LobbyID,
		Players: lobby.Players,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding lobby: %v\n", err)
		http.Error(w, "Failed to encode lobby", http.StatusInternalServerError)
	}
}

// StartFight handles POST requests to /api/lobbies/{lobbyId}/fights
func StartFight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	// Expecting /api/lobbies/{lobbyId}/fights
	if len(pathParts) < 5 || pathParts[4] != "fights" {
		http.Error(w, "Lobby ID is missing or path is malformed", http.StatusBadRequest)
		return
	}
	lobbyID := pathParts[3]

	mutex.Lock()
	defer mutex.Unlock()

	lobby, ok := lobbies[lobbyID]
	if !ok {
		http.Error(w, "Lobby not found", http.StatusNotFound)
		return
	}

	if len(lobby.Players) < 2 { // Assuming a fight needs at least 2 players
		http.Error(w, "Not enough players in lobby to start a fight (requires 2)", http.StatusBadRequest)
		return
	}

	// Retrieve full character data for the first two players in the lobby
	player1Char, p1Exists := characters[lobby.Players[0].CharacterID]
	player2Char, p2Exists := characters[lobby.Players[1].CharacterID]

	if !p1Exists || !p2Exists {
		http.Error(w, "One or more player characters not found", http.StatusInternalServerError) // Should not happen if data is consistent
		return
	}

	fightID := uuid.New().String()
	player1Data := game.InitializeFightPlayer(player1Char)
	player2Data := game.InitializeFightPlayer(player2Char)

	newFight := &models.Fight{
		FightID:             fightID,
		Player1:             player1Data,
		Player2:             player2Data,
		Turn:                1,
		IsOver:              false,
		CurrentTurnPlayerID: player1Data.Character.ID, // Player 1 starts
	}

	activeFights[fightID] = newFight

	// As per README, playerTurn: true if the requester is the one to make the first move.
	// Assuming the one starting the fight is Player1 from the lobby.
	response := map[string]interface{}{
		"fightId":    fightID,
		"playerTurn": true, // Player1 (likely requester) starts
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding start fight response: %v\n", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// MakeMove handles POST requests to /api/fights/{fightId}/moves
func MakeMove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	// Expecting /api/fights/{fightId}/moves
	if len(pathParts) < 5 || pathParts[4] != "moves" {
		http.Error(w, "Fight ID is missing or path is malformed", http.StatusBadRequest)
		return
	}
	fightID := pathParts[3]

	var reqBody struct {
		CharacterID string             `json:"characterId"`
		AttackArea  models.BodyPart    `json:"attackArea"`
		BlockAreas  []models.BodyPart  `json:"blockAreas"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(reqBody.BlockAreas) > game.MaxBlockAreas {
		http.Error(w, fmt.Sprintf("Cannot block more than %d areas", game.MaxBlockAreas), http.StatusBadRequest)
		return
	}
    // Validate BodyPart values
    validBodyParts := map[models.BodyPart]bool{models.Head:true, models.Chest:true, models.Groin:true, models.Legs:true}
    if !validBodyParts[reqBody.AttackArea] {
        http.Error(w, "Invalid attack area specified", http.StatusBadRequest)
        return
    }
    for _, area := range reqBody.BlockAreas {
        if !validBodyParts[area] {
            http.Error(w, "Invalid block area specified", http.StatusBadRequest)
            return
        }
    }


	mutex.Lock()
	defer mutex.Unlock()

	fight, ok := activeFights[fightID]
	if !ok {
		http.Error(w, "Fight not found", http.StatusNotFound)
		return
	}

	if fight.IsOver {
		http.Error(w, "Fight is already over", http.StatusBadRequest)
		return
	}

	if reqBody.CharacterID != fight.CurrentTurnPlayerID {
		http.Error(w, "Not your turn", http.StatusForbidden)
		return
	}

	var attacker, defender *models.FightPlayerData
	if reqBody.CharacterID == fight.Player1.Character.ID {
		attacker = &fight.Player1
		defender = &fight.Player2
	} else if reqBody.CharacterID == fight.Player2.Character.ID {
		attacker = &fight.Player2
		defender = &fight.Player1
	} else {
		http.Error(w, "Character ID does not match any player in this fight", http.StatusBadRequest)
		return
	}

	// Store player's move
	attacker.ChosenAttackArea = reqBody.AttackArea
	attacker.ChosenBlockAreas = reqBody.BlockAreas

	// AI Opponent's move
	defender.ChosenAttackArea = game.ChooseRandomAttackArea()
	defender.ChosenBlockAreas = game.ChooseRandomBlockAreas(game.MaxBlockAreas)

	// Resolve Player's Attack
	attackerStats := game.FighterStats{Strength: attacker.Character.Strength, Agility: attacker.Character.Agility}
	defenderStats := game.FighterStats{Strength: defender.Character.Strength, Agility: defender.Character.Agility}
	
	playerDamage, playerCrit, playerDodge, playerAttackMsg := game.ResolveTurn(attackerStats, defenderStats, attacker.ChosenAttackArea, defender.ChosenBlockAreas)
	defender.CurrentHealth -= playerDamage

	// Resolve Opponent's Attack
	opponentDamage, opponentCrit, opponentDodge, opponentAttackMsg := game.ResolveTurn(defenderStats, attackerStats, defender.ChosenAttackArea, attacker.ChosenBlockAreas)
	attacker.CurrentHealth -= opponentDamage
	
	turnResultMessages := []string{playerAttackMsg, opponentAttackMsg}

	// Check for Fight End
	if attacker.CurrentHealth <= 0 {
		fight.IsOver = true
		fight.WinnerID = defender.Character.ID
		attacker.CurrentHealth = 0 // Ensure health doesn't go negative in response
	} else if defender.CurrentHealth <= 0 {
		fight.IsOver = true
		fight.WinnerID = attacker.Character.ID
		defender.CurrentHealth = 0 // Ensure health doesn't go negative in response
	}

	// Update Turn
	if !fight.IsOver {
		if fight.CurrentTurnPlayerID == fight.Player1.Character.ID {
			fight.CurrentTurnPlayerID = fight.Player2.Character.ID
		} else {
			fight.CurrentTurnPlayerID = fight.Player1.Character.ID
		}
		fight.Turn++
	}
	
	// Update fight in map (important if not using pointers, but we are)
	activeFights[fightID] = fight

	response := map[string]interface{}{
		"playerHit":           playerDamage > 0,
		"opponentHit":         opponentDamage > 0,
		"playerDamageDealt":   playerDamage,
		"opponentDamageDealt": opponentDamage,
		"playerHealth":        attacker.CurrentHealth,
		"opponentHealth":      defender.CurrentHealth,
		"turnResultMessages":  turnResultMessages,
		"isFightOver":         fight.IsOver,
		"winnerId":            fight.WinnerID, // Will be empty if fight not over
		"nextTurnPlayerId":    fight.CurrentTurnPlayerID, // Inform client whose turn is next
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding make move response: %v\n", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// EndFight handles DELETE requests to /api/fights/{fightId}
// Based on README, this is more about getting results and cleaning up.
func EndFight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete { // Or GET if it's just fetching results
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	// Expecting /api/fights/{fightId}
	if len(pathParts) < 4 || pathParts[3] == "" {
		http.Error(w, "Fight ID is missing in URL path", http.StatusBadRequest)
		return
	}
	fightID := pathParts[3]

	mutex.Lock()
	defer mutex.Unlock()

	fight, ok := activeFights[fightID]
	if !ok {
		http.Error(w, "Fight not found", http.StatusNotFound)
		return
	}

	// The README implies this endpoint is called after the fight is over.
	// If it's not over, current behavior is to return results anyway and remove.
	// Consider adding logic if it must be over.

	var winnerIdentifier string
	var actualWinnerCharacterID string // To store the ID of the winning character
	xpAwarded := 0

	if fight.IsOver {
		if fight.WinnerID == fight.Player1.Character.ID {
			winnerIdentifier = fight.Player1.Character.Name
			actualWinnerCharacterID = fight.Player1.Character.ID
		} else if fight.WinnerID == fight.Player2.Character.ID {
			winnerIdentifier = fight.Player2.Character.Name
			actualWinnerCharacterID = fight.Player2.Character.ID
		} else {
			winnerIdentifier = "Draw" // Or handle no winner if that's possible
		}

		if actualWinnerCharacterID != "" {
			xpAwarded = 100 // Fixed XP for now
			
			// Award XP to the winning character
			if winningCharacter, charExists := characters[actualWinnerCharacterID]; charExists {
				// AwardXP needs a pointer, so we need to update the map entry directly
				// or re-assign if characters map stores values (it stores values).
				// To modify the character in the map, we need to get a copy, modify it, then put it back.
				// Or, if AwardXP modifies a pointer, we'd need to ensure characters map stores pointers.
				// Let's assume characters map stores models.Character (values).
				// The game.AwardXP function expects a *models.Character.
				
				// Create a temporary modifiable copy
				tempChar := winningCharacter 
				leveledUp, pointsGained := game.AwardXP(&tempChar, xpAwarded)
				characters[actualWinnerCharacterID] = tempChar // Put the updated character back
				
				if leveledUp {
					fmt.Printf("Character %s (ID: %s) leveled up! Gained %d attribute points.\n", tempChar.Name, tempChar.ID, pointsGained)
				}
			} else {
				fmt.Printf("Error: Winning character ID %s not found in characters map.\n", actualWinnerCharacterID)
			}
		}
	} else {
		// Fight is not over, but request is to end it.
		// Respond with current state, or an error, or force end it.
		// For now, let's say it's not a typical scenario to call DELETE unless fight is over.
		// However, we will still remove it from activeFights as per DELETE request.
		winnerIdentifier = "Fight not concluded" // No XP awarded if fight is not concluded
		xpAwarded = 0
	}
	
	delete(activeFights, fightID) // Remove fight from active map

	response := map[string]interface{}{
		"winner":     winnerIdentifier,
		"experience": xpAwarded, // This will be 0 if fight was not concluded or it was a draw
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding end fight response: %v\n", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
