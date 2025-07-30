<script lang="ts">
  import { Player, type SmallBoard } from './types';

  export let board: SmallBoard;
  export let boardIndex: number;
  export let isActive: boolean;
  export let winner: Player;
  export let onCellClick: (bigBoardIndex: number, smallBoardIndex: number) => void;

  function getPlayerSymbol(player: Player): string {
    switch (player) {
      case Player.X: return 'X';
      case Player.O: return 'O';
      default: return '';
    }
  }

  function getWinnerSymbol(): string {
    return getPlayerSymbol(winner);
  }
</script>

<div class="small-board" class:active={isActive} class:won={winner !== Player.Empty}>
  {#if winner !== Player.Empty}
    <div class="winner-overlay">
      <span class="winner-symbol">{getWinnerSymbol()}</span>
    </div>
  {/if}
  
  <div class="grid">
    {#each board as cell, cellIndex}
      <button
        class="cell"
        class:x={cell === Player.X}
        class:o={cell === Player.O}
        disabled={cell !== Player.Empty || winner !== Player.Empty || !isActive}
        on:click={() => onCellClick(boardIndex, cellIndex)}
      >
        {getPlayerSymbol(cell)}
      </button>
    {/each}
  </div>
</div>

<style>
  .small-board {
    position: relative;
    border: 2px solid #666;
    border-radius: 8px;
    padding: 4px;
    background: #f9f9f9;
    transition: all 0.2s ease;
  }

  .small-board.active {
    border-color: #007acc;
    box-shadow: 0 0 10px rgba(0, 122, 204, 0.3);
    background: #f0f8ff;
  }

  .small-board.won {
    background: #e8e8e8;
    opacity: 0.7;
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    grid-template-rows: repeat(3, 1fr);
    gap: 2px;
    width: 120px;
    height: 120px;
  }

  .cell {
    background: white;
    border: 1px solid #ccc;
    border-radius: 4px;
    font-size: 20px;
    font-weight: bold;
    cursor: pointer;
    transition: background-color 0.2s ease;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .cell:hover:not(:disabled) {
    background: #e6f3ff;
  }

  .cell:disabled {
    cursor: not-allowed;
  }

  .cell.x {
    color: #d32f2f;
  }

  .cell.o {
    color: #1976d2;
  }

  .winner-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(255, 255, 255, 0.9);
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 6px;
    z-index: 1;
  }

  .winner-symbol {
    font-size: 60px;
    font-weight: bold;
    color: #333;
    text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
  }
</style>