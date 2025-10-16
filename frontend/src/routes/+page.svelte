<script lang="ts">
  import Game from '$lib/Game.svelte';
  import MultiplayerGame from '$lib/MultiplayerGame.svelte';
  import AIGame from '$lib/AIGame.svelte';

  let gameMode = $state<'menu' | 'single' | 'multi' | 'ai'>('menu');

  function selectSinglePlayer() {
    gameMode = 'single';
  }

  function selectMultiplayer() {
    gameMode = 'multi';
  }

  function selectAI() {
    gameMode = 'ai';
  }

  function backToMenu() {
    gameMode = 'menu';
  }
</script>

{#if gameMode === 'menu'}
  <div class="menu-container">
    <header class="menu-header">
      <h1>🎯 Ultimate Tic-Tac-Toe</h1>
      <p class="subtitle">Choose your game mode</p>
    </header>

    <div class="menu-options">
      <button class="mode-button single" onclick={selectSinglePlayer}>
        <div class="mode-icon">👥</div>
        <h2>Local 2-Player</h2>
        <p>Play locally on the same device</p>
      </button>

      <button class="mode-button ai" onclick={selectAI}>
        <div class="mode-icon">🤖</div>
        <h2>VS AI</h2>
        <p>Challenge the computer opponent</p>
      </button>

      <button class="mode-button multi" onclick={selectMultiplayer}>
        <div class="mode-icon">🌐</div>
        <h2>Online Multiplayer</h2>
        <p>Play online with a friend</p>
      </button>
    </div>
  </div>
{:else if gameMode === 'single'}
  <div class="game-wrapper">
    <button class="back-button" onclick={backToMenu}>← Back to Menu</button>
    <Game />
  </div>
{:else if gameMode === 'ai'}
  <div class="game-wrapper">
    <button class="back-button" onclick={backToMenu}>← Back to Menu</button>
    <AIGame />
  </div>
{:else if gameMode === 'multi'}
  <div class="game-wrapper">
    <button class="back-button" onclick={backToMenu}>← Back to Menu</button>
    <MultiplayerGame />
  </div>
{/if}

<style>
  .menu-container {
    max-width: 600px;
    margin: 0 auto;
    padding: 40px 20px;
    text-align: center;
    font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
  }

  .menu-header h1 {
    font-size: 3rem;
    color: #333;
    margin-bottom: 10px;
  }

  .subtitle {
    font-size: 1.2rem;
    color: #666;
    margin-bottom: 40px;
  }

  .menu-options {
    display: flex;
    gap: 20px;
    justify-content: center;
    flex-wrap: wrap;
  }

  .mode-button {
    background: white;
    border: 3px solid #e0e0e0;
    border-radius: 16px;
    padding: 25px 20px;
    width: 220px;
    cursor: pointer;
    transition: all 0.3s ease;
    text-align: center;
  }

  .mode-button:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
  }

  .mode-button.single:hover {
    border-color: #ff9800;
    background: #fff8f0;
  }

  .mode-button.ai:hover {
    border-color: #ff9800;
    background: #fff8f0;
  }

  .mode-button.multi:hover {
    border-color: #2196f3;
    background: #f0f8ff;
  }

  .mode-icon {
    font-size: 3rem;
    margin-bottom: 15px;
  }

  .mode-button h2 {
    font-size: 1.5rem;
    margin: 15px 0 10px 0;
    color: #333;
  }

  .mode-button p {
    color: #666;
    font-size: 1rem;
    margin: 0;
  }

  .game-wrapper {
    position: relative;
  }

  .back-button {
    position: absolute;
    top: 20px;
    left: 20px;
    background: #f5f5f5;
    border: 1px solid #ddd;
    padding: 10px 15px;
    border-radius: 6px;
    cursor: pointer;
    font-size: 14px;
    color: #666;
    transition: all 0.2s ease;
  }

  .back-button:hover {
    background: #e0e0e0;
    color: #333;
  }

  @media (max-width: 600px) {
    .menu-options {
      flex-direction: column;
      align-items: center;
    }
    
    .mode-button {
      width: 280px;
    }
    
    .menu-options {
      gap: 15px;
    }
  }
</style>
