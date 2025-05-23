import React, { createContext, useReducer, useContext, ReactNode } from 'react';

// 1. Define State and Action Types
interface CharacterStats {
  strength: number;
  dexterity: number;
  intelligence: number;
  vitality: number;
}

interface CharacterState {
  name: string;
  level: number;
  xp: number;
  stats: CharacterStats;
  availableStatPoints: number;
}

type CharacterAction =
  | { type: 'SET_CHARACTER'; payload: CharacterState }
  | { type: 'LEVEL_UP' }
  | { type: 'ADD_XP'; payload: number }
  | { type: 'UPDATE_STAT'; payload: { stat: keyof CharacterStats; value: number } }
  | { type: 'SET_NAME'; payload: string };

// 2. Define Initial State
const initialCharacterState: CharacterState = {
  name: 'Gladiator',
  level: 1,
  xp: 0,
  stats: {
    strength: 5,
    dexterity: 5,
    intelligence: 5,
    vitality: 5,
  },
  availableStatPoints: 0,
};

// 3. Create Context
interface CharacterContextProps {
  characterState: CharacterState;
  dispatchCharacterAction: React.Dispatch<CharacterAction>;
}

const CharacterContext = createContext<CharacterContextProps | undefined>(undefined);

// 4. Implement Reducer
const characterReducer = (state: CharacterState, action: CharacterAction): CharacterState => {
  switch (action.type) {
    case 'SET_CHARACTER':
      return action.payload;
    case 'SET_NAME':
      return { ...state, name: action.payload };
    case 'ADD_XP':
      // Basic XP and leveling logic (can be expanded)
      const newXp = state.xp + action.payload;
      const xpToNextLevel = state.level * 100; // Example: 100 XP for level 1, 200 for level 2
      if (newXp >= xpToNextLevel) {
        return {
          ...state,
          level: state.level + 1,
          xp: newXp - xpToNextLevel,
          availableStatPoints: state.availableStatPoints + 5, // Example: 5 points per level
        };
      }
      return { ...state, xp: newXp };
    case 'LEVEL_UP': // Directly level up and grant points
        return {
            ...state,
            level: state.level + 1,
            availableStatPoints: state.availableStatPoints + 5, // Grant 5 stat points on level up
        };
    case 'UPDATE_STAT':
      if (state.availableStatPoints > 0) {
        return {
          ...state,
          stats: {
            ...state.stats,
            [action.payload.stat]: state.stats[action.payload.stat] + action.payload.value,
          },
          availableStatPoints: state.availableStatPoints - action.payload.value, // Assuming 1 point per value increase
        };
      }
      return state; // Or handle error: not enough points
    default:
      return state;
  }
};

// 5. Implement Context Provider
interface CharacterProviderProps {
  children: ReactNode;
}

export const CharacterProvider: React.FC<CharacterProviderProps> = ({ children }) => {
  const [characterState, dispatchCharacterAction] = useReducer(characterReducer, initialCharacterState);

  return (
    <CharacterContext.Provider value={{ characterState, dispatchCharacterAction }}>
      {children}
    </CharacterContext.Provider>
  );
};

// 6. Create Custom Hook
export const useCharacter = (): CharacterContextProps => {
  const context = useContext(CharacterContext);
  if (!context) {
    throw new Error('useCharacter must be used within a CharacterProvider');
  }
  return context;
};
