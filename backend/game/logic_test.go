package game

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestCreateAICharacter(t *testing.T) {
	t.Run("Level 1 AI Character", func(t *testing.T) {
		playerLevel := 1
		aiChar := CreateAICharacter(playerLevel)

		if _, err := uuid.Parse(aiChar.ID); err != nil {
			t.Errorf("AI Character ID is not a valid UUID: %v", err)
		}
		if aiChar.ID == "" {
			t.Errorf("Expected AI Character ID to be non-empty, got empty string")
		}

		expectedName := fmt.Sprintf("Computer Gladiator Lv. %d", playerLevel)
		if aiChar.Name != expectedName {
			t.Errorf("Expected AI Character Name to be '%s', got '%s'", expectedName, aiChar.Name)
		}

		if aiChar.Level != playerLevel {
			t.Errorf("Expected AI Character Level to be %d, got %d", playerLevel, aiChar.Level)
		}

		if aiChar.XP != 0 {
			t.Errorf("Expected AI Character XP to be 0, got %d", aiChar.XP)
		}

		if aiChar.AvailablePoints != 0 {
			t.Errorf("Expected AI Character AvailablePoints to be 0, got %d", aiChar.AvailablePoints)
		}

		// Expected stats for Level 1: Strength=8, Agility=8, Stamina=8
		expectedStrength := 8
		if aiChar.Strength != expectedStrength {
			t.Errorf("Expected AI Character Strength to be %d, got %d", expectedStrength, aiChar.Strength)
		}
		expectedAgility := 8
		if aiChar.Agility != expectedAgility {
			t.Errorf("Expected AI Character Agility to be %d, got %d", expectedAgility, aiChar.Agility)
		}
		expectedStamina := 8
		if aiChar.Stamina != expectedStamina {
			t.Errorf("Expected AI Character Stamina to be %d, got %d", expectedStamina, aiChar.Stamina)
		}
	})

	t.Run("Level 5 AI Character", func(t *testing.T) {
		playerLevel := 5
		aiChar := CreateAICharacter(playerLevel)

		if _, err := uuid.Parse(aiChar.ID); err != nil {
			t.Errorf("AI Character ID is not a valid UUID: %v", err)
		}
		if aiChar.ID == "" {
			t.Errorf("Expected AI Character ID to be non-empty, got empty string")
		}


		expectedName := fmt.Sprintf("Computer Gladiator Lv. %d", playerLevel)
		if aiChar.Name != expectedName {
			t.Errorf("Expected AI Character Name to be '%s', got '%s'", expectedName, aiChar.Name)
		}

		if aiChar.Level != playerLevel {
			t.Errorf("Expected AI Character Level to be %d, got %d", playerLevel, aiChar.Level)
		}

		if aiChar.XP != 0 {
			t.Errorf("Expected AI Character XP to be 0, got %d", aiChar.XP)
		}

		if aiChar.AvailablePoints != 0 {
			t.Errorf("Expected AI Character AvailablePoints to be 0, got %d", aiChar.AvailablePoints)
		}

		// Expected stats for Level 5:
		// Strength = 8 + (5 - 1) * 2 = 8 + 4 * 2 = 8 + 8 = 16
		// Agility  = 8 + (5 - 1) * 2 = 8 + 4 * 2 = 8 + 8 = 16
		// Stamina  = 8 + (5 - 1) * 1 = 8 + 4 * 1 = 8 + 4 = 12
		expectedStrength := 16
		if aiChar.Strength != expectedStrength {
			t.Errorf("Expected AI Character Strength to be %d, got %d", expectedStrength, aiChar.Strength)
		}
		expectedAgility := 16
		if aiChar.Agility != expectedAgility {
			t.Errorf("Expected AI Character Agility to be %d, got %d", expectedAgility, aiChar.Agility)
		}
		expectedStamina := 12
		if aiChar.Stamina != expectedStamina {
			t.Errorf("Expected AI Character Stamina to be %d, got %d", expectedStamina, aiChar.Stamina)
		}
	})
}
