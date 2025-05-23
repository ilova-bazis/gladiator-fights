package game

import (
	"fmt"
	"gladiator-backend/models"
	"math/rand"
	"time"

	"github.com/google/uuid" // Added for AI character ID generation
)

// Game balance constants
const (
	BaseDamagePerStrength          = 1
	BaseHealthPerStamina           = 5
	CritMultiplier                 = 1.5
	BaseDodgeChancePerAgilityPoint = 0.02 // 2%
	BaseCritChancePerAgilityPoint  = 0.02 // 2%
	AgilityDiminishingReturnThreshold = 20
	DiminishedDodgeChancePerAgilityPoint = 0.01 // 1%
	DiminishedCritChancePerAgilityPoint  = 0.01 // 1%
	MaxBlockAreas                  = 2
	AttributePointsPerLevel        = 5 // New constant
)

// xpPerLevel defines the XP needed to advance from the current level to the next.
// Key: current level, Value: XP needed to reach current level + 1.
var xpPerLevel = map[int]int{
	1: 100, // To reach level 2
	2: 200, // To reach level 3
	3: 300, // To reach level 4
	4: 400, // To reach level 5
	5: 500, // To reach level 6
	// Add more levels as needed
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// CalculateBaseDamage calculates damage based on strength.
func CalculateBaseDamage(strength int) int {
	return strength * BaseDamagePerStrength
}

// ApplyDiminishingReturns calculates a chance with diminishing returns.
func ApplyDiminishingReturns(statValue int, baseChancePerPoint float64, threshold int, diminishedChancePerPoint float64) float64 {
	if statValue <= 0 {
		return 0.0
	}
	if statValue <= threshold {
		return float64(statValue) * baseChancePerPoint
	}
	return (float64(threshold) * baseChancePerPoint) + (float64(statValue-threshold) * diminishedChancePerPoint)
}

// CalculateDodgeChance calculates dodge chance based on agility using diminishing returns.
func CalculateDodgeChance(agility int) float64 {
	// Example: Agility 25, Base 2% up to 20, Diminished 1% after 20
	// (20 * 0.02) + (5 * 0.01) = 0.40 + 0.05 = 0.45 (45%)
	chance := ApplyDiminishingReturns(agility, BaseDodgeChancePerAgilityPoint, AgilityDiminishingReturnThreshold, DiminishedDodgeChancePerAgilityPoint)
	// Max dodge chance can be capped if needed, e.g., if chance > 0.75 { return 0.75 }
	return chance
}

// CalculateCritChance calculates critical hit chance based on agility using diminishing returns.
func CalculateCritChance(agility int) float64 {
	// Example: Agility 25, Base 2% up to 20, Diminished 1% after 20
	// (20 * 0.02) + (5 * 0.01) = 0.40 + 0.05 = 0.45 (45%)
	chance := ApplyDiminishingReturns(agility, BaseCritChancePerAgilityPoint, AgilityDiminishingReturnThreshold, DiminishedCritChancePerAgilityPoint)
	// Max crit chance can be capped if needed
	return chance
}

// FighterStats defines the necessary stats for a fighter in ResolveTurn.
// Using a specific struct for this to make ResolveTurn more testable and decoupled.
type FighterStats struct {
	Strength int
	Agility  int
	// CurrentHealth is not directly used by ResolveTurn for calculations,
	// but damageDealt will be subtracted from it by the caller.
}

// ResolveTurn processes a single attack exchange between an attacker and a defender.
// It determines the outcome of an attack based on chosen actions and stats.
func ResolveTurn(attacker FighterStats, defender FighterStats, attackArea models.BodyPart, blockAreas []models.BodyPart) (damageDealt int, wasCrit bool, wasDodge bool, message string) {
	// 1. Process Block
	for _, blockedArea := range blockAreas {
		if attackArea == blockedArea {
			return 0, false, false, fmt.Sprintf("Attack on %s was blocked!", attackArea)
		}
	}

	// 2. Process Dodge
	defenderDodgeChance := CalculateDodgeChance(defender.Agility)
	if rand.Float64() < defenderDodgeChance {
		return 0, false, true, "Attack was dodged!"
	}

	// 3. Calculate Damage
	baseDamage := CalculateBaseDamage(attacker.Strength)
	actualDamage := baseDamage // Initialize with base damage

	// 4. Process Critical Hit
	attackerCritChance := CalculateCritChance(attacker.Agility)
	if rand.Float64() < attackerCritChance {
		actualDamage = int(float64(actualDamage) * CritMultiplier)
		wasCrit = true
	}

	damageDealt = actualDamage
	msg := fmt.Sprintf("Attack hit %s for %d damage.", attackArea, damageDealt)
	if wasCrit {
		msg += " Critical hit!"
	}
	return damageDealt, wasCrit, false, msg // wasDodge is false if not dodged
}

// InitializeFightPlayer creates FightPlayerData from a Character model.
// It sets the initial health based on stamina.
func InitializeFightPlayer(character models.Character) models.FightPlayerData {
	initialHealth := character.Stamina * BaseHealthPerStamina
	return models.FightPlayerData{
		Character:       character,
		CurrentHealth:   initialHealth,
		// ChosenAttackArea and ChosenBlockAreas will be set per turn by player input or AI
	}
}

// AllBodyParts is a slice containing all possible body parts for random selection.
var AllBodyParts = []models.BodyPart{models.Head, models.Chest, models.Groin, models.Legs}

// ChooseRandomAttackArea selects a random body part to attack.
func ChooseRandomAttackArea() models.BodyPart {
	return AllBodyParts[rand.Intn(len(AllBodyParts))]
}

// ChooseRandomBlockAreas selects a specified number of unique random body parts to block.
func ChooseRandomBlockAreas(count int) []models.BodyPart {
	if count >= len(AllBodyParts) {
		return AllBodyParts // Return all if requested count is too high
	}
	if count <= 0 {
		return []models.BodyPart{}
	}

	blocked := make(map[models.BodyPart]bool)
	var result []models.BodyPart

	for len(result) < count {
		area := AllBodyParts[rand.Intn(len(AllBodyParts))]
		if !blocked[area] {
			blocked[area] = true
			result = append(result, area)
		}
	}
	return result
}

// AwardXP adds XP to a character and handles leveling up.
// It returns true if the character leveled up, and the number of attribute points gained.
func AwardXP(character *models.Character, amount int) (leveledUp bool, pointsGained int) {
	if character == nil {
		return false, 0
	}

	character.XP += amount
	leveledUp = false
	pointsGained = 0

	// Loop to handle multiple level-ups
	for {
		xpNeeded, ok := xpPerLevel[character.Level]
		if !ok {
			// Max level reached according to the xpPerLevel map or level not defined.
			// If you want to cap XP at max level, you could set character.XP = 0 here
			// or ensure xpNeeded for the max level is very high / effectively infinite.
			// For now, we assume the map covers all progression.
			// If a level isn't in the map, they can't progress further via this mechanism.
			break 
		}

		if character.XP >= xpNeeded {
			leveledUp = true
			character.XP -= xpNeeded
			character.Level++
			character.AvailablePoints += AttributePointsPerLevel
			pointsGained += AttributePointsPerLevel
		} else {
			// Not enough XP for the current level's threshold
			break
		}
	}
	return leveledUp, pointsGained
}

// CreateAICharacter generates an AI opponent character.
// The AI's stats scale with the provided playerLevel.
func CreateAICharacter(playerLevel int) models.Character {
	aiStrength := 8 + (playerLevel-1)*2
	aiAgility := 8 + (playerLevel-1)*2
	aiStamina := 8 + (playerLevel-1)*1

	// Ensure stats are not below a minimum baseline, especially if playerLevel could be < 1 (though current logic handles level 1 okay)
	if aiStrength < 1 {
		aiStrength = 1
	}
	if aiAgility < 1 {
		aiAgility = 1
	}
	if aiStamina < 1 {
		aiStamina = 1
	}

	return models.Character{
		ID:              uuid.New().String(),
		Name:            fmt.Sprintf("Computer Gladiator Lv. %d", playerLevel),
		Level:           playerLevel,
		Strength:        aiStrength,
		Agility:         aiAgility,
		Stamina:         aiStamina,
		XP:              0,
		AvailablePoints: 0,
	}
}
