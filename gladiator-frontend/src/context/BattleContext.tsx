import React, { createContext, useReducer, useContext, ReactNode } from 'react';

// 1. Define State and Action Types
interface BattleParticipant {
  id: string;
  name: string;
  health: number;
  maxHealth: number;
  // Add other relevant stats like attack, defense, speed etc.
}

interface BattleState {
  battleId: string | null;
  player: BattleParticipant | null;
  opponent: BattleParticipant | null;
  currentTurn: 'player' | 'opponent' | null;
  turnLog: string[]; // To store messages like "Player attacks Opponent for X damage"
  isLoading: boolean;
  error: string | null;
  isBattleOver: boolean;
  winner: 'player' | 'opponent' | null;
}

type BattleAction =
  | { type: 'START_BATTLE'; payload: { battleId: string; player: BattleParticipant; opponent: BattleParticipant } }
  | { type: 'PLAYER_ATTACK'; payload: { damageDealt: number; log: string } }
  | { type: 'OPPONENT_ATTACK'; payload: { damageDealt: number; log: string } }
  | { type: 'END_TURN' }
  | { type: 'SET_BATTLE_LOADING'; payload: boolean }
  | { type: 'SET_BATTLE_ERROR'; payload: string | null }
  | { type: 'PROCESS_MOVE_RESULT'; payload: { playerHealth: number; opponentHealth: number; log: string[]; turn: 'player' | 'opponent' | null } }
  | { type: 'END_BATTLE'; payload: { winner: 'player' | 'opponent' | null } }
  | { type: 'ADD_LOG_MESSAGE'; payload: string }
  | { type: 'RESET_BATTLE' };

// 2. Define Initial State
const initialBattleState: BattleState = {
  battleId: null,
  player: null,
  opponent: null,
  currentTurn: null,
  turnLog: [],
  isLoading: false,
  error: null,
  isBattleOver: false,
  winner: null,
};

// 3. Create Context
interface BattleContextProps {
  battleState: BattleState;
  dispatchBattleAction: React.Dispatch<BattleAction>;
}

const BattleContext = createContext<BattleContextProps | undefined>(undefined);

// 4. Implement Reducer
const battleReducer = (state: BattleState, action: BattleAction): BattleState => {
  switch (action.type) {
    case 'START_BATTLE':
      return {
        ...initialBattleState,
        battleId: action.payload.battleId,
        player: action.payload.player,
        opponent: action.payload.opponent,
        currentTurn: 'player', // Player usually starts
        isLoading: false,
        isBattleOver: false,
        winner: null,
        turnLog: [`Battle started between ${action.payload.player.name} and ${action.payload.opponent.name}!`],
      };
    case 'PLAYER_ATTACK':
      if (!state.opponent || !state.player) return state;
      const newOpponentHealth = state.opponent.health - action.payload.damageDealt;
      const playerAttackLog = [...state.turnLog, action.payload.log];
      if (newOpponentHealth <= 0) {
        return {
          ...state,
          opponent: { ...state.opponent, health: 0 },
          turnLog: [...playerAttackLog, `${state.opponent.name} has been defeated!`],
          isBattleOver: true,
          winner: 'player',
          currentTurn: null,
        };
      }
      return {
        ...state,
        opponent: { ...state.opponent, health: newOpponentHealth },
        turnLog: playerAttackLog,
        currentTurn: 'opponent', // Switch turn
      };
    case 'OPPONENT_ATTACK':
      if (!state.player || !state.opponent) return state;
      const newPlayerHealth = state.player.health - action.payload.damageDealt;
      const opponentAttackLog = [...state.turnLog, action.payload.log];
      if (newPlayerHealth <= 0) {
        return {
          ...state,
          player: { ...state.player, health: 0 },
          turnLog: [...opponentAttackLog, `${state.player.name} has been defeated!`],
          isBattleOver: true,
          winner: 'opponent',
          currentTurn: null,
        };
      }
      return {
        ...state,
        player: { ...state.player, health: newPlayerHealth },
        turnLog: opponentAttackLog,
        currentTurn: 'player', // Switch turn
      };
    case 'END_TURN': // This action might be used if there are effects or decisions between turns
        if (!state.isBattleOver) {
            return {
                ...state,
                currentTurn: state.currentTurn === 'player' ? 'opponent' : 'player',
            };
        }
        return state;
    case 'SET_BATTLE_LOADING':
      return { ...state, isLoading: action.payload };
    case 'SET_BATTLE_ERROR':
      return { ...state, error: action.payload, isLoading: false };
    case 'ADD_LOG_MESSAGE':
      return {
        ...state,
        turnLog: [...state.turnLog, action.payload],
      };
    case 'PROCESS_MOVE_RESULT':
      if (!state.player || !state.opponent) return state;
      return {
        ...state,
        player: { ...state.player, health: action.payload.playerHealth },
        opponent: { ...state.opponent, health: action.payload.opponentHealth },
        turnLog: [...state.turnLog, ...action.payload.log],
        currentTurn: action.payload.turn, // Backend should tell us whose turn it is now
      };
    case 'END_BATTLE':
        return {
            ...state,
            isBattleOver: true,
            winner: action.payload.winner,
            currentTurn: null, // No more turns
            isLoading: false,
        };
    case 'RESET_BATTLE':
      return initialBattleState;
    default:
      return state;
  }
};

// 5. Implement Context Provider
interface BattleProviderProps {
  children: ReactNode;
}

export const BattleProvider: React.FC<BattleProviderProps> = ({ children }) => {
  const [battleState, dispatchBattleAction] = useReducer(battleReducer, initialBattleState);

  return (
    <BattleContext.Provider value={{ battleState, dispatchBattleAction }}>
      {children}
    </BattleContext.Provider>
  );
};

// 6. Create Custom Hook
export const useBattle = (): BattleContextProps => {
  const context = useContext(BattleContext);
  if (!context) {
    throw new Error('useBattle must be used within a BattleProvider');
  }
  return context;
};
