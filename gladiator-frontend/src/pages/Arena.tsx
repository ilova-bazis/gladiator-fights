import React, { useState, useEffect, useCallback } from 'react';
import { useCharacter } from '../context/CharacterContext';
import { useBattle } from '../context/BattleContext';
import { useNotification } from '../context/NotificationContext';
import {
  createLobby,
  joinLobby,
  startFight,
  makeMove,
  endFight,
  FightParticipant, // Re-using this from api.ts
  ApiCharacter, // For character details
  MoveResponse
} from '../services/api';
import { Link, useNavigate } from 'react-router-dom';
import './Arena.css'; // We'll create this file for styling

const BODY_AREAS = ['Head', 'Chest', 'Groin', 'Legs'];

const Arena: React.FC = () => {
  const { characterState, dispatchCharacterAction } = useCharacter();
  const { battleState, dispatchBattleAction } = useBattle();
  const { addNotification } = useNotification();
  const navigate = useNavigate();

  const [selectedAttackArea, setSelectedAttackArea] = useState<string | null>(null);
  const [selectedBlockAreas, setSelectedBlockAreas] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(false); // For initial setup
  const [isSubmittingMove, setIsSubmittingMove] = useState<boolean>(false);

  // Opponent's last move for display
  const [opponentLastAttack, setOpponentLastAttack] = useState<string | null>(null);
  const [opponentLastBlocks, setOpponentLastBlocks] = useState<string[]>([]);


  // Helper to get character ID, ensuring it's a number if present
  const getCharacterId = (): number | null => {
    const id = (characterState as any).id;
    return typeof id === 'number' ? id : null;
  };

  const initializeBattle = useCallback(async (charId: number) => {
    setIsLoading(true);
    addNotification('Finding a worthy opponent...', 'info');
    try {
      const lobby = await createLobby();
      addNotification(`Lobby ${lobby.id} created. Joining...`, 'info');

      // JoinLobby might return initial fight data or just confirmation
      // For now, assuming it might give us some details, but startFight is the main source of fight data
      await joinLobby(lobby.id, charId);
      addNotification(`Joined lobby ${lobby.id}. Starting fight...`, 'info');

      const fightData = await startFight(lobby.id);
      addNotification(`Fight ${fightData.id} started!`, 'success', 2000);

      // Map ApiCharacter stats from fightData.player to BattleContext's player structure
      // Backend uses agility and stamina. Frontend context uses dexterity and vitality.
      const mapParticipant = (apiParticipant: FightParticipant, charContext?: ApiCharacter): FightParticipant => ({
        ...apiParticipant,
        // If it's the player, use their detailed stats from CharacterContext for maxHealth calculation,
        // as startFight might only return current health based on base stats, not modified by items/buffs later.
        // However, for simplicity now, we use what backend gives, assuming it's correctly calculated.
        // Max health should be based on vitality/stamina.
        // Let's assume backend's FightParticipant already has correct maxHealth.
        // If not, we'd calculate it: (charContext ? charContext.stamina : apiParticipant.stamina) * 10 (example factor)
      });


      dispatchBattleAction({
        type: 'START_BATTLE',
        payload: {
          battleId: fightData.id,
          player: mapParticipant(fightData.player, characterState as ApiCharacter),
          opponent: mapParticipant(fightData.opponent),
          // currentTurn: fightData.current_turn, // Assuming backend provides this
        },
      });
    } catch (error: any) {
      console.error('Battle initialization failed:', error);
      addNotification(error.message || 'Failed to initialize battle. Please try again.', 'error', 5000);
      // Consider navigating away or allowing a retry
      navigate('/profile'); // Or back to lobby selection if that existed
    } finally {
      setIsLoading(false);
    }
  }, [dispatchBattleAction, addNotification, characterState, navigate]);

  useEffect(() => {
    const charId = getCharacterId();
    if (charId && !battleState.battleId && !isLoading) { // Only initialize if character exists and no battle ongoing
      initializeBattle(charId);
    }
    // Cleanup function to end fight if component unmounts unexpectedly
    // return () => {
    //   if (battleState.battleId && !battleState.isBattleOver) {
    //     // Consider if this is the desired behavior. Maybe only if player initiated leaving.
    //     // endFight(battleState.battleId).catch(err => console.error("Failed to cleanup fight on unmount", err));
    //     // dispatchBattleAction({ type: 'RESET_BATTLE' });
    //   }
    // };
  }, [characterState, battleState.battleId, initializeBattle, isLoading]); // Removed battleState.isBattleOver and dispatchBattleAction from deps of this useEffect

  const handleAttackAreaSelect = (area: string) => {
    setSelectedAttackArea(area);
  };

  const handleBlockAreaSelect = (area: string) => {
    setSelectedBlockAreas((prev) => {
      if (prev.includes(area)) {
        return prev.filter((a) => a !== area);
      }
      if (prev.length < 2) {
        return [...prev, area];
      }
      // If 2 already selected, replace the first one (or provide other UX)
      addNotification('You can only select up to 2 block areas. Deselect one first or this will replace the oldest selection.', 'info', 2000);
      return [prev[1], area]; // Replace the first selected
    });
  };

  const handleSubmitMove = async () => {
    if (!selectedAttackArea) {
      addNotification('Please select an area to attack.', 'error');
      return;
    }
    if (selectedBlockAreas.length !== 2) {
      addNotification('Please select exactly two areas to block.', 'error');
      return;
    }
    if (!battleState.battleId || battleState.currentTurn !== 'player' || battleState.isBattleOver) {
      addNotification('Cannot make a move right now.', 'error');
      return;
    }

    // Log player's chosen actions before making the move
    const playerActionLog = `You chose to attack ${selectedAttackArea} and block ${selectedBlockAreas.join(' & ')}.`;
    dispatchBattleAction({ type: 'ADD_LOG_MESSAGE', payload: playerActionLog });

    setIsSubmittingMove(true);
    try {
      const moveResult = await makeMove(battleState.battleId, selectedAttackArea, selectedBlockAreas);

      // Update opponent's last move for UI display from moveResult
      if (moveResult.opponentAttackArea) setOpponentLastAttack(moveResult.opponentAttackArea);
      if (moveResult.opponentBlockAreas) setOpponentLastBlocks(moveResult.opponentBlockAreas);

      // Process the results of the turn using the new context action
      dispatchBattleAction({
        type: 'PROCESS_MOVE_RESULT',
        payload: {
          playerHealth: moveResult.playerHealth,
          opponentHealth: moveResult.opponentHealth,
          log: moveResult.log,
          turn: moveResult.gameOver ? null : 'player', // If game over, no turn. Else, player's turn again.
                                                       // This assumes backend handles AI turn and returns control to player.
        },
      });

      if (moveResult.gameOver) {
        addNotification(`Fight over! Winner: ${moveResult.winner || 'N/A'}`, 'success', 3000);
        dispatchBattleAction({ type: 'END_BATTLE', payload: { winner: moveResult.winner || null } });
        if (moveResult.xpGained && moveResult.xpGained > 0) {
          dispatchCharacterAction({ type: 'ADD_XP', payload: moveResult.xpGained });
          addNotification(`You gained ${moveResult.xpGained} XP!`, 'success', 3000);
        }
        // No explicit endFight API call here as backend signals game over.
        // If backend requires explicit call even on natural end, it would be here.
      } else {
         // It's opponent's turn now, but we don't need to do anything in UI until it's player's turn again
         // The backend will process AI move and then it will be player's turn.
         // For now, the UI just waits. A more advanced system might show "Opponent is thinking..."
         // and then automatically query for the result of opponent's turn.
         // Given current structure, makeMove likely returns after AI has also moved if it's a quick AI.
         // Or, currentTurn is updated by makeMove response.
         // Let's assume makeMove response indicates new currentTurn or if it's still player's turn (e.g. opponent dodged and has counter)
         // For now, the existing PLAYER_ATTACK/OPPONENT_ATTACK should handle turn switching.
      }

      // Reset selections for next turn
      setSelectedAttackArea(null);
      setSelectedBlockAreas([]);

    } catch (error: any) {
      console.error('Failed to make move:', error);
      addNotification(error.message || 'Failed to make move.', 'error');
    } finally {
      setIsSubmittingMove(false);
    }
  };

  const handlePlayAgain = () => {
    dispatchBattleAction({ type: 'RESET_BATTLE' }); // Reset battle context
    setSelectedAttackArea(null);
    setSelectedBlockAreas([]);
    setOpponentLastAttack(null);
    setOpponentLastBlocks([]);
    const charId = getCharacterId();
    if (charId) {
        initializeBattle(charId); // Start a new battle
    } else {
        addNotification("Cannot start a new game, character data missing.", "error");
        navigate('/');
    }
  };


  // Render checks
  const charId = getCharacterId();
  if (!charId && !isLoading) {
    return (
      <div className="arena-container" style={{ textAlign: 'center', padding: '50px' }}>
        <h2>Arena</h2>
        <p>Please create or select a character first to enter the Arena.</p>
        <Link to="/" className="arena-button">Go to Character Builder</Link>
      </div>
    );
  }

  // NOTE: Arena.css was not converted to a CSS module, so class names are global.
  // We will use the existing kebab-case names from Arena.css. If it were a module,
  // we'd use styles.arenaContainer, etc.

  if (isLoading && !battleState.battleId) {
    return <div className="arenaContainer"><h1>Preparing for Battle...</h1><p>Please wait while we find an opponent.</p></div>;
  }

  if (!battleState.battleId || !battleState.player || !battleState.opponent) {
    return <div className="arenaContainer"><h1>Error</h1><p>Battle data is missing. Try <button onClick={handlePlayAgain} className="arena-button">starting a new game</button> or go to <Link to="/profile" className="arena-button">Profile</Link>.</p></div>;
  }
  
  const HealthBar = ({ current, max }: { current: number; max: number }) => {
    const percentage = max > 0 ? (current / max) * 100 : 0;
    return (
      <div className="health-bar-container">
        <div className="health-bar" style={{ width: `${percentage}%`, backgroundColor: percentage > 60 ? 'green' : percentage > 30 ? 'orange' : 'red' }}>
          {current} / {max}
        </div>
      </div>
    );
  };

  const ParticipantDisplay = ({ participant, isPlayer, lastAttack, lastBlocks }: { participant: FightParticipant, isPlayer: boolean, lastAttack?: string | null, lastBlocks?: string[] }) => (
    <div className={`participant-section ${isPlayer ? 'player' : 'opponent'}`}>
      <div className="participantIcon" style={{ backgroundImage: isPlayer ? `url('../assets/player1_icon.jpeg')` : `url('../assets/player2_icon.jpeg')` }}></div>
      <h3>{participant.name} {isPlayer ? "(You)" : "(Opponent)"}</h3>
      <HealthBar current={participant.health} max={participant.maxHealth} />
      <p>Str: {participant.strength}, Dex: {participant.agility}, Vit: {participant.stamina}</p>
      { !isPlayer && lastAttack && <p style={{color: 'red', fontWeight: 'bold'}}>Last Attack: {lastAttack}</p> }
      { !isPlayer && lastBlocks && lastBlocks.length > 0 && <p style={{color: 'blue', fontWeight: 'bold'}}>Last Blocks: {lastBlocks.join(', ')}</p> }
    </div>
  );

  if (battleState.isBattleOver) {
    return (
      <div className="arenaContainer" style={{ textAlign: 'center' }}>
        <h1>Fight Over!</h1>
        <h2>{battleState.winner === 'player' ? 'You are Victorious!' : battleState.winner === 'opponent' ? 'You have been Defeated.' : 'The battle ended in a draw.'}</h2>
        <div className="battle-log"> {/* Use battle-log class for consistency in log display */}
            {battleState.turnLog.map((log, index) => <p key={index} className={log.includes(battleState.player?.name || "Player") ? 'player-log' : 'opponent-log'}>{log}</p>)}
        </div>
        <button onClick={handlePlayAgain} className="arena-button">Play Again</button>
        <Link to="/profile" className="arena-button">View Profile</Link>
      </div>
    );
  }

  return (
    <div className="arenaContainer">
      <h1 style={{ textAlign: 'center' }}>Arena - Fight {battleState.battleId}</h1>
      <div style={{ textAlign: 'center', marginBottom: '20px', fontWeight: 'bold', fontSize: '1.2em', color: '#fff' }}>
        Turn: {battleState.currentTurn === 'player' ? "Your Turn" : "Opponent's Turn"}
      </div>

      <div className="battle-ground">
        <ParticipantDisplay participant={battleState.player} isPlayer={true} />
        <div className="vs-separator">VS</div>
        <ParticipantDisplay participant={battleState.opponent} isPlayer={false} lastAttack={opponentLastAttack} lastBlocks={opponentLastBlocks} />
      </div>

      {battleState.currentTurn === 'player' && !battleState.isBattleOver && (
        <div className="actions-container">
          <div className="action-section">
            <h4>Choose Attack Area (Select 1)</h4>
            <div className="body-parts">
              {BODY_AREAS.map((area) => (
                <button
                  key={`attack-${area}`}
                  className={`body-part-button ${selectedAttackArea === area ? 'selected-attack' : ''}`}
                  onClick={() => handleAttackAreaSelect(area)}
                >
                  {area}
                </button>
              ))}
            </div>
          </div>

          <div className="action-section">
            <h4>Choose Block Areas (Select 2)</h4>
            <div className="body-parts">
              {BODY_AREAS.map((area) => (
                <button
                  key={`block-${area}`}
                  className={`body-part-button ${selectedBlockAreas.includes(area) ? 'selected-block' : ''}`}
                  onClick={() => handleBlockAreaSelect(area)}
                  disabled={selectedBlockAreas.length >= 2 && !selectedBlockAreas.includes(area)}
                >
                  {area}
                </button>
              ))}
            </div>
          </div>
           <button 
            onClick={handleSubmitMove} 
            disabled={isSubmittingMove || selectedAttackArea === null || selectedBlockAreas.length !== 2}
            className="arena-button submit-move-button"
            >
            {isSubmittingMove ? 'Submitting...' : 'Submit Move'}
          </button>
        </div>
      )}

      <div className="battle-log-container">
        <h4>Battle Log</h4>
        <div className="battle-log">
          {battleState.turnLog.map((log, index) => (
            <p key={index} className={log.includes(battleState.player?.name || "Player") ? 'player-log' : 'opponent-log'}>{log}</p>
          ))}
        </div>
      </div>
    </div>
  );
};

export default Arena;
