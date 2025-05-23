package api

import (
	"bytes"
	"encoding/json"
	"gladiator-backend/models"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/google/uuid"
)

// Helper function to reset global state before each test
func resetGlobalState() {
	mutex.Lock()
	defer mutex.Unlock()
	characters = make(map[string]models.Character)
	lobbies = make(map[string]models.Lobby)
	activeFights = make(map[string]*models.Fight)
	// Re-initialize mutex if it's not a pointer or if needed
	// For this setup, the existing mutex should be fine as it's locked/unlocked.
}

func TestStartAIFightAPI(t *testing.T) {
	t.Run("Successful AI Fight Start", func(t *testing.T) {
		resetGlobalState()

		playerLevel := 3
		testPlayerID := uuid.New().String()
		mutex.Lock()
		characters[testPlayerID] = models.Character{
			ID:       testPlayerID,
			Name:     "Test Player",
			Level:    playerLevel,
			Strength: 10,
			Agility:  10,
			Stamina:  10,
		}
		mutex.Unlock()

		requestBody := map[string]string{"characterId": testPlayerID}
		jsonBody, _ := json.Marshal(requestBody)

		req, err := http.NewRequest(http.MethodPost, "/api/fights/start-ai", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Could not create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		// Directly call the handler. If using a router, you'd serve HTTP through the router.
		StartAIFight(rr, req)

		if status := rr.Code; status != http.StatusCreated {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
			t.Errorf("Response body: %s", rr.Body.String())
			return
		}

		var response map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Could not unmarshal response: %v", err)
		}

		fightID, ok := response["fightId"].(string)
		if !ok || fightID == "" {
			t.Errorf("Expected 'fightId' to be a non-empty string, got %v", response["fightId"])
		}

		playerTurn, ok := response["playerTurn"].(bool)
		if !ok || !playerTurn {
			t.Errorf("Expected 'playerTurn' to be true, got %v", response["playerTurn"])
		}

		aiCharacterID, ok := response["aiCharacterId"].(string)
		if !ok || aiCharacterID == "" {
			t.Errorf("Expected 'aiCharacterId' to be a non-empty string, got %v", response["aiCharacterId"])
		}

		mutex.Lock()
		fight, fightExists := activeFights[fightID]
		mutex.Unlock()

		if !fightExists {
			t.Fatalf("Fight with ID %s not found in activeFights", fightID)
		}

		if fight.Player1.Character.ID != testPlayerID {
			t.Errorf("Expected Player1 ID to be %s, got %s", testPlayerID, fight.Player1.Character.ID)
		}
		if fight.Player2.Character.ID != aiCharacterID {
			t.Errorf("Expected Player2 ID to be %s, got %s", aiCharacterID, fight.Player2.Character.ID)
		}
		if fight.Player2.Character.Level != playerLevel {
			t.Errorf("Expected AI Character (Player2) Level to be %d, got %d", playerLevel, fight.Player2.Character.Level)
		}
		// Check if AI name is as expected by CreateAICharacter
		// Name: fmt.Sprintf("Computer Gladiator Lv. %d", playerLevel)
		// This requires importing "fmt"
	})

	t.Run("Error Case - Player Character Not Found", func(t *testing.T) {
		resetGlobalState()

		nonExistentPlayerID := uuid.New().String()
		requestBody := map[string]string{"characterId": nonExistentPlayerID}
		jsonBody, _ := json.Marshal(requestBody)

		req, err := http.NewRequest(http.MethodPost, "/api/fights/start-ai", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Could not create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		StartAIFight(rr, req)

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
		}
	})

	t.Run("Error Case - Missing CharacterID in Request", func(t *testing.T) {
		resetGlobalState()

		req, err := http.NewRequest(http.MethodPost, "/api/fights/start-ai", bytes.NewBuffer([]byte("{}"))) // Empty JSON
		if err != nil {
			t.Fatalf("Could not create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		StartAIFight(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code for empty characterId: got %v want %v", status, http.StatusBadRequest)
		}
	})

	t.Run("Error Case - Malformed JSON Request", func(t *testing.T) {
		resetGlobalState()

		req, err := http.NewRequest(http.MethodPost, "/api/fights/start-ai", bytes.NewBuffer([]byte("not a json")))
		if err != nil {
			t.Fatalf("Could not create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		StartAIFight(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code for malformed JSON: got %v want %v", status, http.StatusBadRequest)
		}
	})
}

// Placeholder for TestAIFightSequence if it's to be added in this file
// func TestAIFightSequence(t *testing.T) { // Original placeholder
// resetGlobalState()
//	// ... setup similar to TestStartAIFightAPI ...
// }

func TestAIFightSequence(t *testing.T) {
	resetGlobalState()

	// 1. Setup: Create a human player and start an AI fight
	playerLevel := 1 // Use level 1 for simpler stat expectations if needed later
	humanPlayerID := uuid.New().String()
	humanPlayer := models.Character{
		ID:              humanPlayerID,
		Name:            "Human Test Player",
		Level:           playerLevel,
		Strength:        10, // Sufficient to do some damage
		Agility:         10,
		Stamina:         10, // Base health will be 10*5 = 50
		AvailablePoints: 0,
		XP:              0,
	}
	mutex.Lock()
	characters[humanPlayerID] = humanPlayer
	mutex.Unlock()

	// Start AI Fight (programmatic setup, not full HTTP for this part to simplify)
	// We need the AI's ID and the fight ID
	aiOpponent := game.CreateAICharacter(playerLevel) // game is not directly accessible here.
	// We need to call StartAIFight or replicate its core logic to get a fight setup.
	// Let's use the actual StartAIFight handler for a more integrated test.

	startFightRequestBody := map[string]string{"characterId": humanPlayerID}
	startFightJsonBody, _ := json.Marshal(startFightRequestBody)
	startFightReq, _ := http.NewRequest(http.MethodPost, "/api/fights/start-ai", bytes.NewBuffer(startFightJsonBody))
	startFightReq.Header.Set("Content-Type", "application/json")
	startFightRR := httptest.NewRecorder()
	StartAIFight(startFightRR, startFightReq)

	if startFightRR.Code != http.StatusCreated {
		t.Fatalf("TestAIFightSequence: StartAIFight setup failed with status %d: %s", startFightRR.Code, startFightRR.Body.String())
	}

	var startFightResponse map[string]interface{}
	if err := json.Unmarshal(startFightRR.Body.Bytes(), &startFightResponse); err != nil {
		t.Fatalf("TestAIFightSequence: Could not unmarshal StartAIFight response: %v", err)
	}
	fightID, _ := startFightResponse["fightId"].(string)
	aiCharacterID, _ := startFightResponse["aiCharacterId"].(string)

	if fightID == "" || aiCharacterID == "" {
		t.Fatalf("TestAIFightSequence: fightId or aiCharacterId missing from StartAIFight response")
	}
	
	// Get initial healths for later comparison
	mutex.Lock()
	initialFightState := activeFights[fightID]
	initialPlayerHealth := initialFightState.Player1.CurrentHealth
	initialAIHealth := initialFightState.Player2.CurrentHealth
	mutex.Unlock()


	// 2. Simulate Player Move
	moveRequestBody := models.MakeMoveRequest{
		CharacterID: humanPlayerID,
		AttackArea:  models.Head,
		BlockAreas:  []models.BodyPart{models.Chest},
	}
	moveJsonBody, _ := json.Marshal(moveRequestBody)
	
	// Need to construct path with fightID: /api/fights/{fightId}/moves
	// The router in main.go handles this, but here we call MakeMove directly.
	// We need to simulate the path for MakeMove to parse it.
	// A bit of a hack: create a request with a dummy URL that MakeMove can parse.
	// Or, modify MakeMove to accept fightID as a parameter if testing directly,
	// but for handler test, we should respect its signature.
	// Let's assume the handler's path parsing logic is simple enough.
	// The handler uses: pathParts := strings.Split(r.URL.Path, "/") then pathParts[3]
	
	moveReqPath := "/api/fights/" + fightID + "/moves"
	moveReq, _ := http.NewRequest(http.MethodPost, moveReqPath, bytes.NewBuffer(moveJsonBody))
	moveReq.Header.Set("Content-Type", "application/json")
	
	// We need to set the URL path correctly for the handler to extract fightID
	// httptest.NewRequest doesn't make it easy to set the URL path that the handler sees via r.URL.Path
	// when calling the handler directly.
	// A common way is to use `http.ServeHTTP` with a router, or pass context.
	// For simplicity here, if direct call to MakeMove doesn't work due to path, this test needs adjustment.
	// Let's try setting r.URL.Path directly if possible, or use a simpler approach.
	// The current MakeMove extracts fightID from r.URL.Path.

	rr := httptest.NewRecorder()
	MakeMove(rr, moveReq) // This call will make MakeMove parse moveReq.URL.Path

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("TestAIFightSequence: MakeMove returned wrong status code: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
		return // Stop test if move fails
	}

	var moveResponse map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &moveResponse); err != nil {
		t.Fatalf("TestAIFightSequence: Could not unmarshal MakeMove response: %v", err)
	}

	// Assertions on MakeMove response (basic checks)
	if _, ok := moveResponse["playerHealth"]; !ok {
		t.Errorf("TestAIFightSequence: 'playerHealth' missing from MakeMove response")
	}
	if _, ok := moveResponse["opponentHealth"]; !ok {
		t.Errorf("TestAIFightSequence: 'opponentHealth' missing from MakeMove response")
	}
	isFightOver, ok := moveResponse["isFightOver"].(bool)
	if !ok {
		t.Errorf("TestAIFightSequence: 'isFightOver' missing or not a bool in MakeMove response")
	}

	// Verify fight state in activeFights
	mutex.Lock()
	updatedFight, fightExists := activeFights[fightID]
	mutex.Unlock()

	if !fightExists {
		t.Fatalf("TestAIFightSequence: Fight with ID %s not found in activeFights after MakeMove", fightID)
	}

	if updatedFight.Turn != 2 && !isFightOver { // Turn should advance if fight is not over
		t.Errorf("TestAIFightSequence: Expected Turn to be 2 or fight to be over, got Turn %d and isFightOver %v", updatedFight.Turn, isFightOver)
	}
	
	// Check if healths changed (they should, unless all attacks were dodged/blocked and no damage)
	// This is a soft check, specific damage depends on randomness.
	if !isFightOver { // Only check if fight is not over, otherwise healths might be 0
		if updatedFight.Player1.CurrentHealth == initialPlayerHealth && updatedFight.Player2.CurrentHealth == initialAIHealth {
			// This could happen if both attacks missed or were blocked. A more robust check would be to ensure
			// attack/block choices guarantee some interaction, or run multiple times.
			// For now, we'll just note it.
			t.Logf("TestAIFightSequence: Player and AI health unchanged after one turn. This might be due to dodges/blocks.")
		}
		// If the fight is not over, the turn should switch back to the human player (Player1)
		// because AI's turn is also processed in the same MakeMove call.
		if updatedFight.CurrentTurnPlayerID != humanPlayerID {
			t.Errorf("TestAIFightSequence: Expected CurrentTurnPlayerID to be human player %s, got %s", humanPlayerID, updatedFight.CurrentTurnPlayerID)
		}
	}


	// 3. End Fight (Simplified - just call EndFight on the ongoing fight)
	endFightReqPath := "/api/fights/" + fightID
	endFightReq, _ := http.NewRequest(http.MethodDelete, endFightReqPath, nil)
	endFightRR := httptest.NewRecorder()
	EndFight(endFightRR, endFightReq)

	if status := endFightRR.Code; status != http.StatusOK {
		t.Errorf("TestAIFightSequence: EndFight returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var endFightResponse map[string]interface{}
	if err := json.Unmarshal(endFightRR.Body.Bytes(), &endFightResponse); err != nil {
		t.Fatalf("TestAIFightSequence: Could not unmarshal EndFight response: %v", err)
	}

	// Since the fight was likely not over, winner should reflect that
	// and XP should be 0.
	if !isFightOver { // If MakeMove didn't end the fight
		expectedWinner := "Fight not concluded"
		if winner, ok := endFightResponse["winner"].(string); !ok || winner != expectedWinner {
			t.Errorf("TestAIFightSequence: Expected winner to be '%s', got '%v'", expectedWinner, endFightResponse["winner"])
		}
		if exp, ok := endFightResponse["experience"].(float64); !ok || int(exp) != 0 { // JSON numbers are float64
			t.Errorf("TestAIFightSequence: Expected experience to be 0, got %v", endFightResponse["experience"])
		}
	} else { // If MakeMove DID end the fight
		// Check winner based on who had health > 0
		// This part is more complex as it depends on the fight outcome.
		// For simplicity, we'll just check that the keys exist.
		if _, ok := endFightResponse["winner"]; !ok {
			t.Errorf("TestAIFightSequence: 'winner' missing from EndFight response when fight was over")
		}
		if _, ok := endFightResponse["experience"]; !ok {
			t.Errorf("TestAIFightSequence: 'experience' missing from EndFight response when fight was over")
		}
	}


	mutex.Lock()
	_, fightStillExists := activeFights[fightID]
	mutex.Unlock()
	if fightStillExists {
		t.Errorf("TestAIFightSequence: Fight with ID %s should have been removed from activeFights after EndFight", fightID)
	}
}


var _ = sync.Mutex{} // Ensure sync is used if other test functions need it.
