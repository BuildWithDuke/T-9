<script lang="ts">
  import { onMount } from 'svelte';
  import { Player, type GameState, type Move } from './types';
  import SmallBoard from './SmallBoard.svelte';
  import { createGame, getGame, makeMove, makeAIMove } from './api';

  let gameId = $state('');
  let gameState = $state<GameState | null>(null);
  let isLoading = $state(false);
  let error = $state('');
  let difficulty = $state<'easy' | 'medium' | 'hard'>('medium');
  let playerSymbol = $state<Player>(Player.X);
  let isPlayerTurn = $state(true);
  let gameStarted = $state(false);

  // Don't auto-start the game on mount

  async function startNewGame() {
    try {
      isLoading = true;
      error = '';
      
      const response = await createGame();
      gameId = response.gameId;
      gameState = response.game;
      gameStarted = true;
      isPlayerTurn = playerSymbol === Player.X; // X always goes first
      
      // If AI goes first (player chose O), make AI move
      if (!isPlayerTurn) {
        setTimeout(() => makeAIMoveAsync(), 500);
      }
    } catch (err) {
      error = 'Failed to create game. Make sure the backend is running.';
      console.error('Error creating game:', err);
    } finally {
      isLoading = false;
    }
  }

  function resetToSetup() {
    // Reset to setup screen without starting a new game
    gameStarted = false;
    gameState = null;
    gameId = '';
    error = '';
    isLoading = false;
  }

  async function handleCellClick(bigBoardIndex: number, smallBoardIndex: number) {
    if (!gameState || !isPlayerTurn || gameState.gameOver || isLoading) return;

    try {
      isLoading = true;
      error = '';

      const move: Move = {
        bigBoardIndex,
        smallBoardIndex,
        player: playerSymbol
      };

      // Make player move
      const response = await makeMove(gameId, move);
      gameState = response;
      isPlayerTurn = false;

      // If game not over, make AI move after short delay
      if (!gameState.gameOver) {
        setTimeout(() => makeAIMoveAsync(), 800);
      }
    } catch (err) {
      error = 'Failed to make move';
      console.error('Error making move:', err);
      isLoading = false;
    }
  }

  async function makeAIMoveAsync() {
    if (!gameState || gameState.gameOver) return;

    try {
      isLoading = true;
      const response = await makeAIMove(gameId, difficulty);
      gameState = response.game;
      isPlayerTurn = true;
    } catch (err) {
      error = `AI move failed: ${err.message}`;
      console.error('Error making AI move:', err);
      isPlayerTurn = true; // Let player try again
    } finally {
      isLoading = false;
    }
  }

  function changeDifficulty(newDifficulty: 'easy' | 'medium' | 'hard') {
    difficulty = newDifficulty;
    // Don't auto-start, just update the setting
  }

  function changePlayerSymbol(symbol: Player) {
    playerSymbol = symbol;
    // Don't auto-start, just update the setting
  }

  function getPlayerName(player: Player): string {
    switch (player) {
      case Player.X: return 'X';
      case Player.O: return 'O';
      default: return 'Unknown';
    }
  }

  function isSmallBoardActive(boardIndex: number): boolean {
    if (!gameState) return false;
    
    // If board is already won, it's never active
    if (gameState.bigBoardWins[boardIndex] !== Player.Empty) return false;
    
    // If activeBoard is -1, can play in any unwon board
    if (gameState.activeBoard === -1) return true;
    
    // Otherwise, must match the specific active board
    return gameState.activeBoard === boardIndex;
  }

  function getDifficultyColor(diff: string): string {
    switch (diff) {
      case 'easy': return '#4caf50';
      case 'medium': return '#ff9800';
      case 'hard': return '#f44336';
      default: return '#2196f3';
    }
  }

  function getOpponentSymbol(): Player {
    return playerSymbol === Player.X ? Player.O : Player.X;
  }
</script>

<div class="game-container">
  <header class="game-header">
    <h1>🤖 AI Ultimate Tic-Tac-Toe</h1>
    
    {#if !gameStarted}
      <div class="setup-section">
        <h2>Game Setup</h2>
        
        <div class="setting-group">
          <h3>Choose Your Symbol:</h3>
          <div class="symbol-buttons">
            <button 
              class="symbol-btn"
              class:active={playerSymbol === Player.X}
              onclick={() => changePlayerSymbol(Player.X)}
            >
              Play as X (Go First)
            </button>
            <button 
              class="symbol-btn"
              class:active={playerSymbol === Player.O}
              onclick={() => changePlayerSymbol(Player.O)}
            >
              Play as O (Go Second)
            </button>
          </div>
        </div>

        <div class="setting-group">
          <h3>AI Difficulty:</h3>
          <div class="difficulty-buttons">
            <button 
              class="difficulty-btn easy"
              class:active={difficulty === 'easy'}
              onclick={() => changeDifficulty('easy')}
            >
              😊 Easy
            </button>
            <button 
              class="difficulty-btn medium"
              class:active={difficulty === 'medium'}
              onclick={() => changeDifficulty('medium')}
            >
              🤔 Medium
            </button>
            <button 
              class="difficulty-btn hard"
              class:active={difficulty === 'hard'}
              onclick={() => changeDifficulty('hard')}
            >
              🧠 Hard
            </button>
          </div>
        </div>

        <button class="start-btn" onclick={startNewGame} disabled={isLoading}>
          {isLoading ? '⏳ Starting...' : '🚀 Start Game'}
        </button>
      </div>
    {:else}
      <div class="game-info-header">
        <div class="game-status">
          <div class="player-info">
            <span class="you">You: {getPlayerName(playerSymbol)}</span>
            <span class="ai">AI: {getPlayerName(getOpponentSymbol())}</span>
            <span class="difficulty-display" style="color: {getDifficultyColor(difficulty)}">
              {difficulty.toUpperCase()}
            </span>
          </div>
          
          {#if gameState && !gameState.gameOver}
            <div class="turn-indicator">
              {#if isPlayerTurn && !isLoading}
                <span class="your-turn">🎯 Your Turn</span>
              {:else if isLoading}
                <span class="thinking">🤖 AI Thinking...</span>
              {:else}
                <span class="ai-turn">🤖 AI's Turn</span>
              {/if}
            </div>
          {/if}
          
          {#if gameState?.gameOver}
            <div class="game-result">
              {#if gameState.gameWon === playerSymbol}
                🎉 You Win!
              {:else if gameState.gameWon === getOpponentSymbol()}
                🤖 AI Wins!
              {:else}
                🤝 It's a Tie!
              {/if}
            </div>
          {/if}
        </div>
      </div>
    {/if}
  </header>

  {#if error}
    <div class="error">
      ❌ {error}
    </div>
  {/if}

  {#if gameState && gameStarted}
    <div class="big-board" class:disabled={!isPlayerTurn || isLoading}>
      {#each gameState.bigBoard as smallBoard, boardIndex}
        <SmallBoard
          board={smallBoard}
          boardIndex={boardIndex}
          isActive={isSmallBoardActive(boardIndex) && isPlayerTurn && !gameState.gameOver && !isLoading}
          winner={gameState.bigBoardWins[boardIndex]}
          onCellClick={handleCellClick}
        />
      {/each}
    </div>

    <div class="game-info">
      {#if gameState.activeBoard !== -1 && !gameState.gameOver && isPlayerTurn}
        <p>Must play in highlighted board</p>
      {:else if !gameState.gameOver && isPlayerTurn}
        <p>Can play in any available board</p>
      {/if}
    </div>

    <div class="game-controls">
      <button onclick={resetToSetup} disabled={isLoading}>
        🔄 New Game
      </button>
      
      <div class="difficulty-controls">
        <span>Change Difficulty:</span>
        <button 
          class="mini-difficulty-btn easy"
          class:active={difficulty === 'easy'}
          onclick={() => changeDifficulty('easy')}
          disabled={isLoading}
        >
          Easy
        </button>
        <button 
          class="mini-difficulty-btn medium"
          class:active={difficulty === 'medium'}
          onclick={() => changeDifficulty('medium')}
          disabled={isLoading}
        >
          Medium
        </button>
        <button 
          class="mini-difficulty-btn hard"
          class:active={difficulty === 'hard'}
          onclick={() => changeDifficulty('hard')}
          disabled={isLoading}
        >
          Hard
        </button>
      </div>
    </div>
  {/if}
</div>

<style>
  .game-container {
    max-width: 800px;
    margin: 0 auto;
    padding: 20px;
    font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
  }

  .game-header {
    text-align: center;
    margin-bottom: 20px;
  }

  h1 {
    color: #333;
    margin-bottom: 20px;
  }

  .setup-section {
    background: #f8f9fa;
    border-radius: 12px;
    padding: 30px;
    margin: 20px 0;
  }

  .setup-section h2 {
    color: #333;
    margin-bottom: 25px;
  }

  .setting-group {
    margin-bottom: 25px;
  }

  .setting-group h3 {
    color: #555;
    margin-bottom: 15px;
    font-size: 16px;
  }

  .symbol-buttons, .difficulty-buttons {
    display: flex;
    gap: 10px;
    justify-content: center;
    flex-wrap: wrap;
  }

  .symbol-btn, .difficulty-btn {
    padding: 12px 20px;
    border: 2px solid #ddd;
    border-radius: 8px;
    background: white;
    cursor: pointer;
    font-size: 14px;
    font-weight: bold;
    transition: all 0.2s ease;
  }

  .symbol-btn:hover, .difficulty-btn:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }

  .symbol-btn.active {
    background: #007acc;
    color: white;
    border-color: #007acc;
  }

  .difficulty-btn.easy.active {
    background: #4caf50;
    color: white;
    border-color: #4caf50;
  }

  .difficulty-btn.medium.active {
    background: #ff9800;
    color: white;
    border-color: #ff9800;
  }

  .difficulty-btn.hard.active {
    background: #f44336;
    color: white;
    border-color: #f44336;
  }

  .start-btn {
    background: #007acc;
    color: white;
    border: none;
    padding: 15px 30px;
    border-radius: 8px;
    font-size: 16px;
    font-weight: bold;
    cursor: pointer;
    margin-top: 20px;
    transition: background-color 0.2s ease;
  }

  .start-btn:hover:not(:disabled) {
    background: #005a9e;
  }

  .start-btn:disabled {
    background: #ccc;
    cursor: not-allowed;
  }

  .game-info-header {
    margin-bottom: 20px;
  }

  .game-status {
    background: #f8f9fa;
    border-radius: 12px;
    padding: 20px;
    margin: 20px 0;
  }

  .player-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 15px;
    font-weight: bold;
  }

  .you {
    color: #2196f3;
  }

  .ai {
    color: #ff5722;
  }

  .difficulty-display {
    font-size: 14px;
    padding: 4px 8px;
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.8);
  }

  .turn-indicator {
    text-align: center;
    font-size: 18px;
    font-weight: bold;
  }

  .your-turn {
    color: #2196f3;
  }

  .ai-turn {
    color: #ff5722;
  }

  .thinking {
    color: #ff9800;
    animation: pulse 1.5s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.6; }
  }

  .game-result {
    font-size: 24px;
    font-weight: bold;
    color: #2e7d32;
    margin: 20px 0;
    text-align: center;
  }

  .error {
    background: #ffebee;
    color: #c62828;
    padding: 10px;
    border-radius: 4px;
    margin: 10px 0;
    text-align: center;
  }

  .big-board {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    grid-template-rows: repeat(3, 1fr);
    gap: 8px;
    max-width: 400px;
    margin: 20px auto;
    padding: 20px;
    background: #e8e8e8;
    border-radius: 12px;
    transition: opacity 0.3s ease;
  }

  .big-board.disabled {
    opacity: 0.7;
    pointer-events: none;
  }

  .game-info {
    text-align: center;
    margin: 20px 0;
    font-style: italic;
    color: #666;
  }

  .game-controls {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 15px;
    margin-top: 20px;
  }

  .difficulty-controls {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 14px;
  }

  .mini-difficulty-btn {
    padding: 6px 12px;
    border: 1px solid #ddd;
    border-radius: 4px;
    background: white;
    cursor: pointer;
    font-size: 12px;
    transition: all 0.2s ease;
  }

  .mini-difficulty-btn:hover:not(:disabled) {
    background: #f0f0f0;
  }

  .mini-difficulty-btn.easy.active {
    background: #4caf50;
    color: white;
    border-color: #4caf50;
  }

  .mini-difficulty-btn.medium.active {
    background: #ff9800;
    color: white;
    border-color: #ff9800;
  }

  .mini-difficulty-btn.hard.active {
    background: #f44336;
    color: white;
    border-color: #f44336;
  }

  .mini-difficulty-btn:disabled {
    background: #f5f5f5;
    cursor: not-allowed;
    opacity: 0.6;
  }

  button {
    background: #007acc;
    color: white;
    border: none;
    padding: 12px 24px;
    border-radius: 6px;
    font-size: 16px;
    cursor: pointer;
    transition: background-color 0.2s ease;
  }

  button:hover:not(:disabled) {
    background: #005a9e;
  }

  button:disabled {
    background: #ccc;
    cursor: not-allowed;
  }

  @media (max-width: 600px) {
    .symbol-buttons, .difficulty-buttons {
      flex-direction: column;
    }
    
    .player-info {
      flex-direction: column;
      gap: 10px;
    }
    
    .difficulty-controls {
      flex-wrap: wrap;
      justify-content: center;
    }
  }
</style>