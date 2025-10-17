<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Player, type GameState, type Move } from './types';
  import { GameWebSocket, type WebSocketEvents } from './websocket';
  import SmallBoard from './SmallBoard.svelte';

  let gameId = $state('');
  let inputGameId = $state('');
  let showJoinForm = $state(true);
  
  // Reactive state
  let connected = $state(false);
  let gameState = $state<GameState | null>(null);
  let error = $state<string>('');
  let playerJoinMessage = $state<string>('');

  // Create WebSocket client with event handlers
  let wsClient: GameWebSocket;

  onMount(async () => {
    const events: WebSocketEvents = {
      onConnect: () => {
        connected = true;
        error = '';
      },
      onDisconnect: () => {
        connected = false;
      },
      onGameState: (newGameState: GameState) => {
        gameState = newGameState;
      },
      onPlayerJoin: (message: string) => {
        playerJoinMessage = message;
      },
      onError: (errorMessage: string) => {
        error = errorMessage;
      }
    };

    wsClient = new GameWebSocket(undefined, events);

    try {
      await wsClient.connect();
    } catch (err) {
      error = 'Failed to connect to server. Make sure the backend is running on port 8080.';
    }
  });

  onDestroy(() => {
    if (wsClient) {
      wsClient.disconnect();
    }
  });

  function createGame() {
    // Generate a fun fruit/vegetable-based game ID
    const produce = [
      // Fruits
      'apple', 'banana', 'cherry', 'dragon', 'elderberry', 'fig', 'grape', 'honeydew',
      'kiwi', 'lemon', 'mango', 'nectarine', 'orange', 'papaya', 'raspberry',
      'strawberry', 'watermelon', 'blueberry', 'coconut', 'avocado', 'pineapple',
      'peach', 'plum', 'lime', 'grapefruit', 'cantaloupe', 'apricot', 'blackberry',
      // Vegetables
      'broccoli', 'carrot', 'celery', 'cucumber', 'eggplant', 'kale', 'lettuce',
      'mushroom', 'onion', 'potato', 'radish', 'spinach', 'tomato', 'zucchini',
      'pepper', 'corn', 'bean', 'pea', 'cabbage', 'asparagus', 'artichoke', 'beet'
    ];
    
    gameId = produce[Math.floor(Math.random() * produce.length)];
    inputGameId = gameId;
    joinGame();
  }

  function joinGame() {
    if (!inputGameId.trim()) {
      error = 'Please enter a game ID';
      return;
    }

    if (!wsClient.isConnected()) {
      error = 'Not connected to server';
      return;
    }

    gameId = inputGameId.trim();
    wsClient.joinGame(gameId);
    showJoinForm = false;
    error = '';
  }

  function handleCellClick(bigBoardIndex: number, smallBoardIndex: number) {
    if (!gameState || gameState.gameOver || !wsClient.isConnected()) return;

    const move: Move = {
      bigBoardIndex,
      smallBoardIndex,
      player: gameState.currentPlayer
    };

    wsClient.makeMove(move);
  }

  function getPlayerName(player: Player): string {
    switch (player) {
      case Player.X: return 'Player X';
      case Player.O: return 'Player O';
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

  function resetGame() {
    showJoinForm = true;
    gameId = '';
    inputGameId = '';
    gameState = null;
    error = '';
    playerJoinMessage = '';
  }

  function copyGameId() {
    navigator.clipboard.writeText(gameId);
    playerJoinMessage = 'Game ID copied to clipboard!';
    setTimeout(() => playerJoinMessage = '', 2000);
  }
</script>

<div class="game-container">
  <header class="game-header">
    <h1>Ultimate Tic-Tac-Toe - Multiplayer</h1>
    
    <div class="connection-status" class:connected class:disconnected={!connected}>
      {connected ? '🟢 Connected' : '🔴 Disconnected'}
    </div>

    {#if showJoinForm}
      <div class="lobby">
        <h2>Join or Create Game</h2>
        <div class="lobby-controls">
          <button onclick={createGame} disabled={!connected}>
            🎮 Create New Game
          </button>
          <div class="join-form">
            <input 
              bind:value={inputGameId} 
              placeholder="Enter Game ID" 
              onkeydown={(e) => e.key === 'Enter' && joinGame()}
            />
            <button onclick={joinGame} disabled={!connected || !inputGameId.trim()}>
              🚪 Join Game
            </button>
          </div>
        </div>
      </div>
    {:else}
      <div class="game-info-header">
        <div class="game-id-section">
          <h3>🎮 Share this Game ID with your friend:</h3>
          <div class="game-id-display">
            <code class="game-id-code">{gameId}</code>
            <button class="copy-btn" onclick={copyGameId} title="Copy Game ID">📋 Copy</button>
          </div>
          <p class="instruction">Your friend should click "Join Game" and enter this ID</p>
        </div>
        
        {#if gameState && !gameState.gameOver}
          <p class="current-player">
            Current Turn: <span class="player-{gameState.currentPlayer}">{getPlayerName(gameState.currentPlayer)}</span>
          </p>
        {/if}
        
        {#if gameState?.gameOver}
          <p class="game-result">
            {#if gameState.gameWon !== Player.Empty}
              🎉 {getPlayerName(gameState.gameWon)} Wins!
            {:else}
              🤝 It's a Tie!
            {/if}
          </p>
        {/if}
      </div>
    {/if}
  </header>

  {#if error}
    <div class="error">
      ❌ {error}
    </div>
  {/if}

  {#if playerJoinMessage}
    <div class="success">
      ✅ {playerJoinMessage}
    </div>
  {/if}

  {#if gameState && !showJoinForm}
    <div class="big-board">
      {#each gameState.bigBoard as smallBoard, boardIndex}
        <SmallBoard
          board={smallBoard}
          boardIndex={boardIndex}
          isActive={isSmallBoardActive(boardIndex) && !gameState.gameOver}
          winner={gameState.bigBoardWins[boardIndex]}
          onCellClick={handleCellClick}
        />
      {/each}
    </div>

    <div class="game-info">
      {#if gameState.activeBoard !== -1 && !gameState.gameOver}
        <p>Must play in highlighted board</p>
      {:else if !gameState.gameOver}
        <p>Can play in any available board</p>
      {/if}
    </div>

    <div class="game-controls">
      <button onclick={resetGame}>
        🏠 Back to Lobby
      </button>
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
    margin-bottom: 10px;
  }

  .connection-status {
    padding: 8px 16px;
    border-radius: 20px;
    font-weight: bold;
    margin: 10px 0;
    display: inline-block;
  }

  .connection-status.connected {
    background: #e8f5e8;
    color: #2e7d32;
  }

  .connection-status.disconnected {
    background: #ffebee;
    color: #c62828;
  }

  .lobby {
    background: #f5f5f5;
    padding: 30px;
    border-radius: 12px;
    margin: 20px 0;
  }

  .lobby h2 {
    margin-bottom: 20px;
    color: #555;
  }

  .lobby-controls {
    display: flex;
    flex-direction: column;
    gap: 15px;
    align-items: center;
  }

  .join-form {
    display: flex;
    gap: 10px;
    align-items: center;
  }

  .join-form input {
    padding: 10px;
    border: 2px solid #ddd;
    border-radius: 6px;
    font-size: 16px;
    width: 200px;
  }

  .join-form input:focus {
    outline: none;
    border-color: #007acc;
  }

  .game-info-header {
    margin-bottom: 20px;
  }

  .game-id-section {
    background: #f8f9fa;
    border: 2px solid #007acc;
    border-radius: 12px;
    padding: 20px;
    margin: 20px 0;
    text-align: center;
  }

  .game-id-section h3 {
    margin: 0 0 15px 0;
    color: #333;
    font-size: 18px;
  }

  .game-id-display {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    margin: 15px 0;
  }

  .game-id-code {
    background: #fff;
    border: 2px solid #007acc;
    padding: 12px 20px;
    border-radius: 8px;
    font-family: 'Courier New', monospace;
    font-size: 20px;
    font-weight: bold;
    color: #007acc;
    letter-spacing: 2px;
  }

  .copy-btn {
    background: #007acc;
    color: white;
    border: none;
    padding: 12px 16px;
    border-radius: 8px;
    cursor: pointer;
    font-size: 14px;
    font-weight: bold;
    transition: background-color 0.2s ease;
  }

  .copy-btn:hover {
    background: #005a9e;
  }

  .instruction {
    margin: 10px 0 0 0;
    font-size: 14px;
    color: #666;
    font-style: italic;
  }

  .current-player {
    font-size: 18px;
    margin: 10px 0;
  }

  .player-1 { color: #d32f2f; font-weight: bold; }
  .player-2 { color: #1976d2; font-weight: bold; }

  .game-result {
    font-size: 24px;
    font-weight: bold;
    color: #2e7d32;
    margin: 20px 0;
  }

  .error {
    background: #ffebee;
    color: #c62828;
    padding: 10px;
    border-radius: 4px;
    margin: 10px 0;
    text-align: center;
  }

  .success {
    background: #e8f5e8;
    color: #2e7d32;
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
  }

  .game-info {
    text-align: center;
    margin: 20px 0;
    font-style: italic;
    color: #666;
  }

  .game-controls {
    text-align: center;
    margin-top: 20px;
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
</style>