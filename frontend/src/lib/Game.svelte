<script lang="ts">
  import { onMount } from 'svelte';
  import { Player, type GameState, type Move } from './types';
  import { GameAPI } from './api';
  import SmallBoard from './SmallBoard.svelte';

  let gameState: GameState | null = null;
  let gameId: string = '';
  let error: string = '';
  let loading: boolean = false;

  onMount(() => {
    createNewGame();
  });

  async function createNewGame() {
    try {
      loading = true;
      error = '';
      const response = await GameAPI.createGame();
      gameId = response.gameId;
      gameState = response.game;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to create game';
    } finally {
      loading = false;
    }
  }

  async function handleCellClick(bigBoardIndex: number, smallBoardIndex: number) {
    if (!gameState || gameState.gameOver) return;

    try {
      error = '';
      
      const move: Move = {
        bigBoardIndex,
        smallBoardIndex,
        player: gameState.currentPlayer
      };

      const newGameState = await GameAPI.makeMove(gameId, move);
      if (newGameState) {
        gameState = newGameState;
      }
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to make move';
      // Don't nullify the game state on error, keep the current state
    }
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
</script>

<div class="game-container">
  <header class="game-header">
    <h1>Ultimate Tic-Tac-Toe</h1>
    {#if gameState && !gameState.gameOver}
      <p class="current-player">
        Current Player: <span class="player-{gameState.currentPlayer}">{getPlayerName(gameState.currentPlayer)}</span>
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
  </header>

  {#if error}
    <div class="error">
      ❌ {error}
    </div>
  {/if}

  {#if loading}
    <div class="loading">⏳ Loading...</div>
  {/if}

  {#if gameState}
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
      <button onclick={createNewGame} disabled={loading}>
        🎮 New Game
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

  .loading {
    text-align: center;
    font-size: 18px;
    margin: 20px 0;
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