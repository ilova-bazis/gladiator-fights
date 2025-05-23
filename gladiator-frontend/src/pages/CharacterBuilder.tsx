import React, { useState, useEffect, ChangeEvent, FormEvent } from 'react';
import { useCharacter } from '../context/CharacterContext';
import { useNotification } from '../context/NotificationContext';
import { createCharacter, updateCharacterAttributes } from '../services/api'; // ApiCharacter is also used but implicitly via CharacterState
import { useNavigate } from 'react-router-dom'; // For navigation
import styles from './CharacterBuilder.module.css'; // Import CSS Module

const CharacterBuilder: React.FC = () => {
  const { characterState, dispatchCharacterAction } = useCharacter();
  const { addNotification } = useNotification();
  const navigate = useNavigate();

  // Local state for form inputs
  const [name, setName] = useState<string>(characterState.name || 'Gladiator');
  const [strength, setStrength] = useState<number>(characterState.stats.strength);
  const [dexterity, setDexterity] = useState<number>(characterState.stats.dexterity); // Frontend uses dexterity
  const [vitality, setVitality] = useState<number>(characterState.stats.vitality);   // Frontend uses vitality
  const [availablePoints, setAvailablePoints] = useState<number>(characterState.availableStatPoints);

  // Store character ID if it exists (e.g., from context after creation or if loaded)
  const [characterId, setCharacterId] = useState<number | null>(null); // Let's assume ID is part of CharacterState if loaded

  useEffect(() => {
    // If character data is already in context (e.g. user navigated back after creation)
    // And if we imagine the backend returns an 'id' field which we store in context.
    // For now, we'll assume 'characterState.id' might exist if a character was loaded.
    // This part needs to be fleshed out once character loading is implemented.
    if (characterState && (characterState as any).id) {
        setCharacterId((characterState as any).id);
        setName(characterState.name);
        setStrength(characterState.stats.strength);
        setDexterity(characterState.stats.dexterity);
        setVitality(characterState.stats.vitality);
        setAvailablePoints(characterState.availableStatPoints);
    } else {
        // Initialize with default/initial values if no character in context or no ID
        // These values could come from a configuration or be hardcoded defaults for new characters
        const defaultStrength = 10;
        const defaultDexterity = 10;
        const defaultVitality = 10;
        const initialPointsToAllocate = 5; // Points a new character gets to allocate

        setStrength(defaultStrength);
        setDexterity(defaultDexterity);
        setVitality(defaultVitality);
        // availablePoints should be what's left from initial character creation allowance
        // If characterState.level is 1 and no points spent, this would be initialPointsToAllocate
        // If characterState has values, it means it's loaded, otherwise, it's a new char
        if(characterState.level === 1 && characterState.xp === 0 && characterState.availableStatPoints === 0) {
             // This implies a truly new character or a reset character from context default
            setAvailablePoints(initialPointsToAllocate);
             // And set base stats for a fresh character
            dispatchCharacterAction({
                type: 'SET_CHARACTER',
                payload: {
                    ...characterState,
                    name: name, // Keep local name or default 'Gladiator'
                    stats: { strength: defaultStrength, dexterity: defaultDexterity, intelligence: characterState.stats.intelligence, vitality: defaultVitality },
                    availableStatPoints: initialPointsToAllocate,
                },
            });

        } else {
            // Existing character loaded from context, use its available points
            setAvailablePoints(characterState.availableStatPoints);
        }
    }
  }, [characterState, dispatchCharacterAction, name]);


  const handleStatChange = (stat: 'strength' | 'dexterity' | 'vitality', increment: number) => {
    const currentStatValue = stat === 'strength' ? strength : stat === 'dexterity' ? dexterity : vitality;
    if (increment > 0 && availablePoints >= increment) {
      if (stat === 'strength') setStrength(strength + increment);
      else if (stat === 'dexterity') setDexterity(dexterity + increment);
      else if (stat === 'vitality') setVitality(vitality + increment);
      setAvailablePoints(availablePoints - increment);
    } else if (increment < 0 ) {
        // Ensure stat doesn't go below initial base (e.g. 5 or 10, depending on game rules)
        // For this example, let's use the initial characterState stats as base minimum
        const baseStat = characterState.stats[stat]; // This might need adjustment based on how base stats are defined
        if (currentStatValue + increment >= baseStat ) {
            if (stat === 'strength') setStrength(strength + increment);
            else if (stat === 'dexterity') setDexterity(dexterity + increment);
            else if (stat === 'vitality') setVitality(vitality + increment);
            setAvailablePoints(availablePoints - increment); // increment is negative, so this adds points back
        } else {
            addNotification(`Cannot decrease ${stat} below base value.`, 'error');
        }
    } else if (increment > 0 && availablePoints < increment) {
        addNotification('Not enough available stat points.', 'error');
    }
  };

  const getCritChance = (agi: number): number => {
    if (agi <= 20) {
      return agi * 2;
    } else {
      return 20 * 2 + (agi - 20) * 1;
    }
  };

  const getDodgeChance = getCritChance; // Same calculation for dodge
  const getBonusHp = (vit: number): number => vit * 5;
  const getBonusDamage = (str: number): number => str * 1;


  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();

    if (!name.trim()) {
      addNotification('Character name cannot be empty.', 'error');
      return;
    }

    try {
      let currentCharacterId = characterId;

      if (!currentCharacterId) { // Create new character
        addNotification('Creating character...', 'info');
        const newCharacterData = await createCharacter(name);
        currentCharacterId = newCharacterData.id;
        setCharacterId(newCharacterData.id); // Store the ID from backend

        // Update context with basic info from creation (level, xp, etc. might come from backend)
        // The backend uses 'stamina', 'agility'. Frontend uses 'vitality', 'dexterity'.
        dispatchCharacterAction({
          type: 'SET_CHARACTER',
          payload: {
            id: newCharacterData.id, // Make sure id is part of CharacterState
            name: newCharacterData.name,
            level: newCharacterData.level,
            xp: newCharacterData.xp,
            stats: { // Assuming backend returns these directly, adjust if structure differs
              strength: newCharacterData.strength,
              dexterity: newCharacterData.agility, // Map agility from backend
              intelligence: characterState.stats.intelligence, // Keep existing intelligence or default
              vitality: newCharacterData.stamina,     // Map stamina from backend
            },
            availableStatPoints: newCharacterData.available_stat_points, // from backend
          },
        });
        addNotification(`Character "${newCharacterData.name}" created! ID: ${newCharacterData.id}. Now allocating stats.`, 'success');
      }

      if (currentCharacterId) {
        addNotification('Updating character attributes...', 'info');
        // API expects 'stamina' for vitality/endurance and 'agility' for dexterity
        const statsToUpdate = {
          strength: strength,
          agility: dexterity, // map frontend dexterity to backend agility
          stamina: vitality,  // map frontend vitality to backend stamina
        };
        const updatedCharacterData = await updateCharacterAttributes(currentCharacterId, statsToUpdate);

        // Update context with the final stats
        dispatchCharacterAction({
          type: 'SET_CHARACTER',
          payload: {
            id: updatedCharacterData.id,
            name: updatedCharacterData.name,
            level: updatedCharacterData.level,
            xp: updatedCharacterData.xp,
            stats: {
              strength: updatedCharacterData.strength,
              dexterity: updatedCharacterData.agility,
              intelligence: characterState.stats.intelligence, // Or from backend if it sends it
              vitality: updatedCharacterData.stamina,
            },
            // availableStatPoints should be what backend calculated after spending points
            // This assumes the updateCharacterAttributes call deducts points on the backend
            // and returns the new available_stat_points.
            // If backend doesn't auto-deduct and return, then frontend 'availablePoints' state is the source of truth.
            // For now, let's assume backend returns the updated available_stat_points.
            availableStatPoints: updatedCharacterData.available_stat_points,
          },
        });
        addNotification('Character attributes saved successfully!', 'success', 3000);
        // navigate('/profile'); // Optional: navigate to profile page
      }
    } catch (error: any) {
      console.error('Failed to save character:', error);
      addNotification(error.message || 'Failed to save character. Please try again.', 'error');
    }
  };


  return (
    <div className={styles.container}>
      <h1 className={styles.title}>{characterId ? 'Edit Character' : 'Create Your Gladiator'}</h1>
      <form onSubmit={handleSubmit}>
        <div className={styles.formGroup}>
          <label htmlFor="characterName">Character Name:</label>
          <input
            type="text"
            id="characterName"
            value={name}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setName(e.target.value)}
            required
          />
        </div>

        <div className={styles.statsSection}>
          <h3>Stats (Available Points: {availablePoints})</h3>
          {['strength', 'dexterity', 'vitality'].map((statKey) => {
            const statName = statKey.charAt(0).toUpperCase() + statKey.slice(1);
            const currentValue = statKey === 'strength' ? strength : statKey === 'dexterity' ? dexterity : vitality;
            const baseStatValue = characterState.stats[statKey as keyof typeof characterState.stats];
            return (
              <div key={statKey} className={styles.statRow}>
                <span className={styles.statName}>{statName}</span>
                <span className={styles.statValue}>{currentValue}</span>
                <div className={styles.statControls}>
                  <button 
                    type="button" 
                    onClick={() => handleStatChange(statKey as 'strength' | 'dexterity' | 'vitality', -1)}
                    disabled={currentValue <= baseStatValue && availablePoints === characterState.availableStatPoints} // Disable if at base and no points were added from available pool
                  >
                    -
                  </button>
                  <button 
                    type="button" 
                    onClick={() => handleStatChange(statKey as 'strength' | 'dexterity' | 'vitality', 1)}
                    disabled={availablePoints <= 0}
                  >
                    +
                  </button>
                </div>
              </div>
            );
          })}
        </div>

        <div className={styles.statEffects}>
            <h4>Stat Effects:</h4>
            <p>Strength: +{getBonusDamage(strength)} Damage</p>
            <p>Dexterity: {getCritChance(dexterity).toFixed(1)}% Critical Hit Chance, {getDodgeChance(dexterity).toFixed(1)}% Dodge Chance</p>
            <p>Vitality: +{getBonusHp(vitality)} HP</p>
            <small>Dexterity gives +2% per point up to 20, then +1% per point.</small>
        </div>

        <button type="submit" className={styles.submitButton}>
          {characterId ? 'Save Changes' : 'Create Character & Save Stats'}
        </button>
      </form>
    </div>
  );
};

export default CharacterBuilder;
