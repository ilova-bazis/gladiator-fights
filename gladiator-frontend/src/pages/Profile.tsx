import React, { useState, useEffect } from 'react'; // Removed ChangeEvent as it's not used
import { useCharacter } from '../context/CharacterContext';
import { useNotification } from '../context/NotificationContext';
import { updateCharacterAttributes } from '../services/api';
import { Link } from 'react-router-dom';
import styles from './Profile.module.css'; // Import CSS Module

const Profile: React.FC = () => {
  const { characterState, dispatchCharacterAction } = useCharacter();
  const { addNotification } = useNotification();

  // Local state for stat allocation, initialized from context
  const [tempStrength, setTempStrength] = useState(characterState.stats.strength);
  const [tempDexterity, setTempDexterity] = useState(characterState.stats.dexterity);
  const [tempVitality, setTempVitality] = useState(characterState.stats.vitality);
  const [pointsToSpend, setPointsToSpend] = useState(characterState.availableStatPoints);

  // Character ID from context (assuming it's stored after creation/loading)
  // The characterState should ideally have an 'id' field.
  const characterId = (characterState as any).id as number | undefined;

  useEffect(() => {
    // Update local state if context changes (e.g., after API update from elsewhere or initial load)
    setTempStrength(characterState.stats.strength);
    setTempDexterity(characterState.stats.dexterity);
    setTempVitality(characterState.stats.vitality);
    setPointsToSpend(characterState.availableStatPoints);
  }, [characterState]);

  if (!characterId && characterState.name === 'Gladiator' && characterState.level === 1 && characterState.xp === 0) {
    return (
      <div className={styles.noCharacterMessage}>
        <h1>Profile</h1>
        <p>No character created or loaded yet.</p>
        <p>
          Go to the <Link to="/">Character Builder</Link> to create your gladiator.
        </p>
      </div>
    );
  }

  const xpToNextLevel = characterState.level * 100;
  const xpProgressPercent = xpToNextLevel > 0 ? (characterState.xp / xpToNextLevel) * 100 : 0;

  // Derived stat calculation functions (similar to CharacterBuilder)
  const getCritChance = (dex: number): number => {
    if (dex <= 20) return dex * 2;
    return 20 * 2 + (dex - 20) * 1;
  };
  const getDodgeChance = getCritChance;
  const getBonusHp = (vit: number): number => vit * 5;
  const getBonusDamage = (str: number): number => str * 1;

  const handleStatPointChange = (stat: 'strength' | 'dexterity' | 'vitality', increment: number) => {
    const currentStatValue = stat === 'strength' ? tempStrength : stat === 'dexterity' ? tempDexterity : tempVitality;
    const originalStatValue = characterState.stats[stat];

    if (increment > 0 && pointsToSpend >= increment) {
      if (stat === 'strength') setTempStrength(tempStrength + increment);
      else if (stat === 'dexterity') setTempDexterity(tempDexterity + increment);
      else if (stat === 'vitality') setTempVitality(tempVitality + increment);
      setPointsToSpend(pointsToSpend - increment);
    } else if (increment < 0 && currentStatValue + increment >= originalStatValue) {
      // Allow decreasing only down to the original stat value from context (before spending points on this page)
      if (stat === 'strength') setTempStrength(tempStrength + increment);
      else if (stat === 'dexterity') setTempDexterity(tempDexterity + increment);
      else if (stat === 'vitality') setTempVitality(tempVitality + increment);
      setPointsToSpend(pointsToSpend - increment); // increment is negative, adds points back
    } else if (increment > 0 && pointsToSpend < increment) {
      addNotification('Not enough available stat points to spend.', 'error');
    } else if (increment < 0 && currentStatValue + increment < originalStatValue) {
      addNotification(`Cannot decrease ${stat} below its current saved value.`, 'error');
    }
  };

  const handleSaveChanges = async () => {
    if (!characterId) {
      addNotification('Character ID not found. Cannot save changes.', 'error');
      return;
    }

    const pointsActuallySpent = characterState.availableStatPoints - pointsToSpend;
    if (pointsActuallySpent === 0 && 
        tempStrength === characterState.stats.strength &&
        tempDexterity === characterState.stats.dexterity &&
        tempVitality === characterState.stats.vitality) {
      addNotification('No changes made to stats.', 'info');
      return;
    }
    
    // Ensure no stat is lower than its original value from context if points were added
    // This check is mostly covered by handleStatPointChange, but as a safeguard:
    if (tempStrength < characterState.stats.strength || tempDexterity < characterState.stats.dexterity || tempVitality < characterState.stats.vitality) {
        addNotification('Stats cannot be reduced below their current permanent values.', 'error');
        // Reset to context values if an invalid state was somehow reached
        setTempStrength(characterState.stats.strength);
        setTempDexterity(characterState.stats.dexterity);
        setTempVitality(characterState.stats.vitality);
        setPointsToSpend(characterState.availableStatPoints);
        return;
    }


    try {
      addNotification('Saving new stat distribution...', 'info');
      // API expects 'stamina' for vitality and 'agility' for dexterity
      const statsToUpdate = {
        strength: tempStrength,
        agility: tempDexterity,
        stamina: tempVitality,
      };

      const updatedCharacter = await updateCharacterAttributes(characterId, statsToUpdate);

      dispatchCharacterAction({
        type: 'SET_CHARACTER',
        payload: {
          ...characterState, // Preserve other parts of state like ID
          id: updatedCharacter.id,
          name: updatedCharacter.name,
          level: updatedCharacter.level,
          xp: updatedCharacter.xp,
          stats: {
            strength: updatedCharacter.strength,
            dexterity: updatedCharacter.agility, // map from backend agility
            intelligence: characterState.stats.intelligence, // assuming intelligence is not changed here
            vitality: updatedCharacter.stamina,   // map from backend stamina
          },
          availableStatPoints: updatedCharacter.available_stat_points,
        },
      });
      addNotification('Stats updated successfully!', 'success', 3000);
    } catch (error: any) {
      console.error('Failed to update stats:', error);
      addNotification(error.message || 'Failed to update stats. Please try again.', 'error');
      // Revert local changes on error to reflect context state
      setTempStrength(characterState.stats.strength);
      setTempDexterity(characterState.stats.dexterity);
      setTempVitality(characterState.stats.vitality);
      setPointsToSpend(characterState.availableStatPoints);
    }
  };

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>{characterState.name}'s Profile</h1>
      <div className={styles.characterInfo}>
        <p>Level: {characterState.level}</p>
        <p>XP: {characterState.xp} / {xpToNextLevel}</p>
        <div className={styles.xpBarContainer}>
          <div
            className={styles.xpBar}
            style={{ width: `${xpProgressPercent}%` }}
          >
            {Math.round(xpProgressPercent)}%
          </div>
        </div>
      </div>

      <div className={styles.mainContent}>
        <div className={styles.statsDisplay}>
          <h3>Current Stats</h3>
          <p>Strength: {characterState.stats.strength} (+{getBonusDamage(characterState.stats.strength)} Dmg)</p>
          <p>Dexterity: {characterState.stats.dexterity} ({getCritChance(characterState.stats.dexterity).toFixed(1)}% Crit/Dodge)</p>
          <p>Vitality: {characterState.stats.vitality} (+{getBonusHp(characterState.stats.vitality)} HP)</p>
          <p>Intelligence: {characterState.stats.intelligence}</p>
          <hr />
          <p className={styles.availablePointsText}>Available Stat Points: {characterState.availableStatPoints}</p>
        </div>

        {characterState.availableStatPoints > 0 && (
          <div className={styles.allocationSection}>
            <h3>Allocate Points (Available: {pointsToSpend})</h3>
            {['strength', 'dexterity', 'vitality'].map((statKey) => {
              const statName = statKey.charAt(0).toUpperCase() + statKey.slice(1);
              let currentValue, bonusEffect;
              if (statKey === 'strength') {
                currentValue = tempStrength;
                bonusEffect = `+${getBonusDamage(currentValue)} Dmg`;
              } else if (statKey === 'dexterity') {
                currentValue = tempDexterity;
                bonusEffect = `${getCritChance(currentValue).toFixed(1)}% Crit/Dodge`;
              } else { // vitality
                currentValue = tempVitality;
                bonusEffect = `+${getBonusHp(currentValue)} HP`;
              }

              return (
                <div key={statKey} className={styles.statRow}>
                  <span className={styles.statName}>{statName}: {currentValue} ({bonusEffect})</span>
                  <div className={styles.statControls}>
                    <button 
                        onClick={() => handleStatPointChange(statKey as 'strength' | 'dexterity' | 'vitality', -1)}
                        disabled={currentValue <= characterState.stats[statKey as keyof typeof characterState.stats]}
                    >-</button>
                    <button 
                        onClick={() => handleStatPointChange(statKey as 'strength' | 'dexterity' | 'vitality', 1)}
                        disabled={pointsToSpend === 0}
                    >+</button>
                  </div>
                </div>
              );
            })}
            <button 
                onClick={handleSaveChanges} 
                className={styles.saveButton}
                disabled={pointsToSpend === characterState.availableStatPoints}
            >
              Save Stats
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

export default Profile;
