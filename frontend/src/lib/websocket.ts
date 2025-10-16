import type { GameState, Move, Player } from './types';

 // Helper function to get the WebSocket URL based on current location
function getDefaultWebSocketURL(): string {
	if (typeof window === 'undefined') {
		return 'ws://localhost:8080/ws';
	}
	const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
	const host = window.location.host;
	return '${protocol}//${host}/ws';
}

export interface WSMessage {
  type: string;
  gameId?: string;
  move?: Move;
  game?: GameState;
  player?: Player;
  error?: string;
  message?: string;
}

export interface WebSocketEvents {
  onConnect?: () => void;
  onDisconnect?: () => void;
  onGameState?: (gameState: GameState) => void;
  onPlayerJoin?: (message: string) => void;
  onPlayerLeave?: (message: string) => void;
  onError?: (error: string) => void;
  onMove?: (move: Move) => void;
}

export class GameWebSocket {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private events: WebSocketEvents = {};
  private connectionPromise: Promise<void> | null = null;

  // State properties that components can access
  public connected: boolean = false;
  public gameState: GameState | null = null;
  public currentPlayer: Player | null = null;
  public error: string = '';
  public playerJoinMessage: string = '';

  constructor(
    private url: string = import.meta.env.VITE_WS_URL || getDefaultWebSocketURL(),
    events?: WebSocketEvents
  ) {
    if (events) {
      this.events = { ...events };
    }
  }

  connect(): Promise<void> {
    if (this.connectionPromise) {
      return this.connectionPromise;
    }

    this.connectionPromise = new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
          this.connected = true;
          this.reconnectAttempts = 0;
          this.error = '';
          this.events.onConnect?.();
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const message: WSMessage = JSON.parse(event.data);
            this.handleMessage(message);
          } catch (err) {
            console.error('Failed to parse WebSocket message:', err);
            this.events.onError?.('Failed to parse message');
          }
        };

        this.ws.onclose = (event) => {
          this.connected = false;
          this.events.onDisconnect?.();
          this.attemptReconnect();
        };

        this.ws.onerror = (error) => {
          this.error = 'Connection error - backend may not be running';
          this.events.onError?.(this.error);
          reject(error);
        };
      } catch (err) {
        reject(err);
      }
    });

    return this.connectionPromise;
  }

  private handleMessage(message: WSMessage) {
    switch (message.type) {
      case 'game_state':
        if (message.game) {
          this.gameState = message.game;
          this.events.onGameState?.(message.game);
        }
        break;

      case 'player_join':
        if (message.message) {
          this.playerJoinMessage = message.message;
          this.events.onPlayerJoin?.(message.message);
          // Clear message after 3 seconds
          setTimeout(() => {
            this.playerJoinMessage = '';
          }, 3000);
        }
        break;

      case 'player_leave':
        if (message.message) {
          this.error = message.message;
          this.events.onPlayerLeave?.(message.message);
        }
        break;

      case 'error':
        if (message.error) {
          this.error = message.error;
          this.events.onError?.(message.error);
        }
        break;

      case 'move':
        if (message.move) {
          this.events.onMove?.(message.move);
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
      this.events.onError?.(this.error);
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
      this.events.onError?.(this.error);
    }
  }

  private attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      
      setTimeout(() => {
        this.connectionPromise = null;
        this.connect().catch(() => {
          // Connection failed, will try again if under limit
        });
      }, this.reconnectDelay * this.reconnectAttempts);
    } else {
      this.error = 'Failed to reconnect to server';
      this.events.onError?.(this.error);
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.connected = false;
    this.connectionPromise = null;
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  // Method to update event handlers
  updateEvents(events: Partial<WebSocketEvents>) {
    this.events = { ...this.events, ...events };
  }
}
