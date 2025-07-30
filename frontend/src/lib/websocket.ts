import type { GameState, Move, Player } from './types';

export interface WSMessage {
  type: string;
  gameId?: string;
  move?: Move;
  game?: GameState;
  player?: Player;
  error?: string;
  message?: string;
}

export class GameWebSocket {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;

  // State properties that components can access
  public connected: boolean = false;
  public gameState: GameState | null = null;
  public currentPlayer: Player | null = null;
  public error: string = '';
  public playerJoinMessage: string = '';

  constructor(private url: string = 'ws://localhost:8080/ws') {}

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
          this.connected = true;
          this.reconnectAttempts = 0;
          this.error = '';
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const message: WSMessage = JSON.parse(event.data);
            this.handleMessage(message);
          } catch (err) {
            console.error('Failed to parse WebSocket message:', err);
          }
        };

        this.ws.onclose = (event) => {
          this.connected = false;
          this.attemptReconnect();
        };

        this.ws.onerror = (error) => {
          this.error = 'Connection error - backend may not be running';
          reject(error);
        };
      } catch (err) {
        reject(err);
      }
    });
  }

  private handleMessage(message: WSMessage) {
    switch (message.type) {
      case 'game_state':
        if (message.game) {
          this.gameState = message.game;
        }
        break;

      case 'player_join':
        if (message.message) {
          this.playerJoinMessage = message.message;
          // Clear message after 3 seconds
          setTimeout(() => this.playerJoinMessage = '', 3000);
        }
        break;

      case 'player_leave':
        if (message.message) {
          this.error = message.message;
        }
        break;

      case 'error':
        if (message.error) {
          this.error = message.error;
        }
        break;

      default:
        console.log('Unknown message type:', message.type);
    }
  }

  joinGame(gameId: string) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const message: WSMessage = {
        type: 'join_game',
        gameId
      };
      this.ws.send(JSON.stringify(message));
    } else {
      this.error = 'Not connected to server';
    }
  }

  makeMove(move: Move) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const message: WSMessage = {
        type: 'move',
        move
      };
      this.ws.send(JSON.stringify(message));
    } else {
      this.error = 'Not connected to server';
    }
  }

  private attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      
      setTimeout(() => {
        this.connect().catch(() => {
          // Connection failed, will try again if under limit
        });
      }, this.reconnectDelay * this.reconnectAttempts);
    } else {
      this.error = 'Failed to reconnect to server';
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.connected = false;
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}