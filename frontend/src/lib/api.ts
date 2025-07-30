import type { GameState, Move, GameResponse } from './types';

const API_BASE = 'http://localhost:8080/api/v1';

export class GameAPI {
  static async createGame(): Promise<GameResponse> {
    const response = await fetch(`${API_BASE}/games`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    });
    
    if (!response.ok) {
      throw new Error('Failed to create game');
    }
    
    return await response.json();
  }

  static async getGame(gameId: string): Promise<GameState> {
    const response = await fetch(`${API_BASE}/games/${gameId}`);
    
    if (!response.ok) {
      throw new Error('Failed to get game');
    }
    
    return await response.json();
  }

  static async makeMove(gameId: string, move: Move): Promise<GameState> {
    const response = await fetch(`${API_BASE}/games/${gameId}/moves`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(move),
    });
    
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to make move');
    }
    
    return await response.json();
  }
}