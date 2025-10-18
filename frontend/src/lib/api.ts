import type { GameState, Move, GameResponse } from './types';
import { safeFetch, APIError, ErrorFactory, safeAsync } from './error';

const API_BASE = import.meta.env.VITE_API_BASE || 'http://localhost:8080/api/v1';

export interface AIResponse {
  game: GameState;
  move: Move;
}

export class GameAPI {
  static async createGame(): Promise<GameResponse> {
    const result = await safeAsync(async () => {
      const response = await safeFetch(`${API_BASE}/games`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      return await response.json();
    });

    if (!result) {
      throw ErrorFactory.network('Failed to create game');
    }

    return result;
  }

  static async getGame(gameId: string): Promise<GameState> {
    const result = await safeAsync(async () => {
      const response = await safeFetch(`${API_BASE}/games/${gameId}`);
      return await response.json();
    });

    if (!result) {
      throw ErrorFactory.network('Failed to get game');
    }

    return result;
  }

  static async makeMove(gameId: string, move: Move): Promise<GameState> {
    const result = await safeAsync(async () => {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 5000); // 5 second timeout

      try {
        const response = await safeFetch(`${API_BASE}/games/${gameId}/moves`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(move),
          signal: controller.signal,
        });

        clearTimeout(timeoutId);
        return await response.json();
      } catch (error) {
        clearTimeout(timeoutId);
        if (error.name === 'AbortError') {
          throw ErrorFactory.game('Move request timed out');
        }
        throw error;
      }
    });

    if (!result) {
      throw ErrorFactory.game('Failed to make move');
    }

    return result;
  }

  static async makeAIMove(gameId: string, difficulty: 'easy' | 'medium' | 'hard'): Promise<AIResponse> {
    const result = await safeAsync(async () => {
      const response = await safeFetch(`${API_BASE}/games/${gameId}/ai-move`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ difficulty }),
      });

      return await response.json();
    });

    if (!result) {
      throw ErrorFactory.game('Failed to make AI move');
    }

    return result;
  }

  // Health check method
  static async healthCheck(): Promise<boolean> {
    return safeAsync(async () => {
      try {
        const response = await fetch(`${API_BASE}/health`, {
          method: 'GET',
        });
        return response.ok;
      } catch {
        return false;
      }
    }) || false;
  }
}

// Export individual functions for convenience
export const createGame = GameAPI.createGame;
export const getGame = GameAPI.getGame;
export const makeMove = GameAPI.makeMove;
export const makeAIMove = GameAPI.makeAIMove;