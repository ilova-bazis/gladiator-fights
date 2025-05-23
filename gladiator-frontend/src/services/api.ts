// Basic Error structure from the backend (can be expanded)
interface ApiError {
  message: string;
  details?: any;
}

// Character structure from the backend
export interface ApiCharacter { // Exporting for use in components
  id: number;
  name: string;
  level: number;
  xp: number;
  strength: number;
  agility: number; // Backend uses agility for dexterity
  stamina: number; // Backend uses stamina for vitality/endurance
  available_stat_points: number;
  // Add other fields as the backend provides them
}

interface CharacterAttributesUpdate {
  strength: number;
  agility: number;
  stamina: number;
}

// --- Lobby and Fight Types ---
export interface Lobby {
  id: string; // Assuming string ID from backend
  // other lobby details if any
}

export interface FightParticipant { // Similar to BattleParticipant in BattleContext
    id: string; // or number, depending on backend (character ID or a fight-specific ID)
    name: string;
    health: number;
    maxHealth: number;
    // include stats if provided by backend and needed by UI directly here
    strength: number;
    agility: number; // dexterity
    stamina: number; // vitality
}

export interface Fight {
  id: string; // Fight ID
  player: FightParticipant;
  opponent: FightParticipant;
  current_turn: 'player' | 'opponent'; // Or however the backend indicates turn
  // other fight details
}

export interface MoveResponse {
  playerHit: boolean;
  opponentHit: boolean;
  playerDamageDealt: number;
  opponentDamageDealt: number;
  playerHealth: number;
  opponentHealth: number;
  log: string[]; // Backend might send a log of what happened
  gameOver?: boolean;
  winner?: 'player' | 'opponent';
  xpGained?: number;
  // Opponent's move details for display (optional, depends on backend)
  opponentAttackArea?: string;
  opponentBlockAreas?: string[];
}

export interface EndFightResponse {
    message: string;
    xpGained?: number; // XP gained from the fight
}


const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8000/api'; // Replace with your actual API base URL

/**
 * Creates a new character.
 * @param name - The name of the character.
 * @returns A Promise resolving to the created character data.
 * @throws ApiError if the request fails.
 */
export const createCharacter = async (name: string): Promise<ApiCharacter> => {
  const response = await fetch(`${API_BASE_URL}/characters`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ name }),
  });

  if (!response.ok) {
    const errorData: ApiError = await response.json().catch(() => ({ message: 'Failed to create character and parse error' }));
    console.error('API Error creating character:', errorData);
    throw errorData;
  }
  return response.json() as Promise<ApiCharacter>;
};

/**
 * Updates the attributes of an existing character.
 * @param id - The ID of the character to update.
 *   stats - An object containing the new strength, agility, and stamina values.
 * @returns A Promise resolving to the updated character data.
 * @throws ApiError if the request fails.
 */
export const updateCharacterAttributes = async (
  id: number,
  stats: CharacterAttributesUpdate
): Promise<ApiCharacter> => {
  const response = await fetch(`${API_BASE_URL}/characters/${id}/attributes`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(stats),
  });

  if (!response.ok) {
    const errorData: ApiError = await response.json().catch(() => ({ message: 'Failed to update attributes and parse error' }));
    console.error('API Error updating attributes:', errorData);
    throw errorData;
  }
  return response.json() as Promise<ApiCharacter>;
};

// --- Lobby API Calls ---

/**
 * Creates a new lobby.
 * @returns A Promise resolving to the created lobby data.
 */
export const createLobby = async (): Promise<Lobby> => {
  const response = await fetch(`${API_BASE_URL}/lobbies`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
  });
  if (!response.ok) {
    const errorData: ApiError = await response.json().catch(() => ({ message: 'Failed to create lobby' }));
    throw errorData;
  }
  return response.json() as Promise<Lobby>;
};

/**
 * Joins an existing lobby.
 * @param lobbyId - The ID of the lobby to join.
 * @param characterId - The ID of the character joining.
 * @returns A Promise resolving to the lobby join confirmation or fight details.
 */
export const joinLobby = async (lobbyId: string, characterId: number): Promise<any> => { // Adjust 'any' based on actual response
  const response = await fetch(`${API_BASE_URL}/lobbies/${lobbyId}/join`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ character_id: characterId }),
  });
  if (!response.ok) {
    const errorData: ApiError = await response.json().catch(() => ({ message: 'Failed to join lobby' }));
    throw errorData;
  }
  return response.json(); // Might return lobby details or initial fight state
};

// --- Fight API Calls ---

/**
 * Starts a new fight within a lobby.
 * @param lobbyId - The ID of the lobby where the fight will start.
 * @returns A Promise resolving to the initial fight data.
 */
export const startFight = async (lobbyId: string): Promise<Fight> => {
  const response = await fetch(`${API_BASE_URL}/lobbies/${lobbyId}/fights`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
  });
  if (!response.ok) {
    const errorData: ApiError = await response.json().catch(() => ({ message: 'Failed to start fight' }));
    throw errorData;
  }
  return response.json() as Promise<Fight>;
};

/**
 * Makes a move in an ongoing fight.
 * @param fightId - The ID of the fight.
 * @param attackArea - The area the player targets.
 * @param blockAreas - An array of areas the player blocks.
 * @returns A Promise resolving to the outcome of the move.
 */
export const makeMove = async (
  fightId: string,
  attackArea: string,
  blockAreas: string[]
): Promise<MoveResponse> => {
  const response = await fetch(`${API_BASE_URL}/fights/${fightId}/moves`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ attack_area: attackArea, block_areas: blockAreas }),
  });
  if (!response.ok) {
    const errorData: ApiError = await response.json().catch(() => ({ message: 'Failed to make move' }));
    throw errorData;
  }
  return response.json() as Promise<MoveResponse>;
};

/**
 * Ends an ongoing fight (e.g., player flees or an error occurs).
 * This might also be called automatically by the backend when a winner is determined.
 * @param fightId - The ID of the fight to end.
 * @returns A Promise resolving to a confirmation message and any XP gained.
 */
export const endFight = async (fightId: string): Promise<EndFightResponse> => {
  const response = await fetch(`${API_BASE_URL}/fights/${fightId}`, {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
  });
  if (!response.ok) {
    // If the fight ended normally (e.g. health reached 0), backend might return 200 OK with XP
    // and this DELETE is just for cleanup or fleeing.
    // If it's an error during DELETE, then it's an issue.
    // The provided MoveResponse seems to handle gameOver, so this DELETE might be for explicit fleeing.
    const errorData: ApiError = await response.json().catch(() => ({ message: 'Failed to end fight or parse error' }));
    throw errorData;
  }
  // If successful, the response might be empty or contain final confirmation/XP.
  // The MoveResponse already seems to handle XP gain on game over.
  // This needs clarification on when endFight is called and what it returns.
  // For now, assuming it might return XP if called for reasons other than HP depletion handled by makeMove.
  return response.json() as Promise<EndFightResponse>;
};


// Add other API service functions here as needed (e.g., getCharacter, listArenas, etc.)
